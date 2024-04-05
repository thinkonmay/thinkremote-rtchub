package proxy

/*
#define QUEUE_SIZE 16
#define PACKET_SIZE 32 * 1024

enum QueueType {
    Video0,
    Video1,
    Audio,
    Microphone,
    Max
};

typedef struct {
    int is_idr;
    enum QueueType type;
}Metadata;

typedef struct {
    int size;
    Metadata metadata;
    char data[PACKET_SIZE];
} Packet;

typedef enum _EventType {
    POINTER_VISIBLE,
    CHANGE_BITRATE,
    CHANGE_FRAMERATE,
    IDR_FRAME,

    STOP,
    HDR_CALLBACK,
    EVENT_TYPE_MAX
} EventType;

typedef enum _DataType {
    HDR_INFO,
} DataType;

typedef struct {
    int value_number;
    char value_raw[PACKET_SIZE];
    int data_size;

    DataType type;

    int read;
} Event;


typedef struct _Queue{
    Packet array[QUEUE_SIZE];
    int order[QUEUE_SIZE];
}Queue;

typedef struct {
    Queue queues[Max];
    Event events[EVENT_TYPE_MAX];
}SharedMemory;

*/
import "C"
import (
	"errors"
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

type SharedMemory C.SharedMemory

func (memory *SharedMemory) Raise(event_id ,value int) {
	memory.events[event_id].value_number = C.int(value);
	memory.events[event_id].read = 0;
}

const (
	Video0     = C.Video0
	Video1     = C.Video1
	Audio      = C.Audio
	Microphone = C.Microphone
	Max        = C.Max

	POINTER_VISIBLE  = C.POINTER_VISIBLE
	CHANGE_BITRATE   = C.CHANGE_BITRATE
	CHANGE_FRAMERATE = C.CHANGE_FRAMERATE
	IDR_FRAME        = C.IDR_FRAME
	STOP             = C.STOP
	HDR_CALLBACK     = C.HDR_CALLBACK
	EVENT_TYPE_MAX   = C.EVENT_TYPE_MAX
)

const (
	audio = 1
	video = 2
)

type DataType int

func peek(memory *C.SharedMemory, media DataType) bool {
	if media == video {
		return memory.queues[C.Video0].order[0] != -1
	} else if media == audio {
		return memory.queues[C.Audio].order[0] != -1
	}

	panic(fmt.Errorf("unknown data type"))
}

func ObtainSharedMemory(token string) (*SharedMemory,error) {
	mod, err := syscall.LoadDLL("libparent.dll")
	if err != nil {
		return nil, err
	}
	obtain, err := mod.FindProc("obtain_shared_memory")
	if err != nil {
		return nil, err
	}
	lock, err := mod.FindProc("lock_shared_memory")
	if err != nil {
		return nil, err
	}
	unlock, err := mod.FindProc("unlock_shared_memory")
	if err != nil {
		return nil, err
	}

	buffer := []byte(token)
	pointer, _, err := obtain.Call(
		uintptr(unsafe.Pointer(&buffer[0])),
	)
	if !errors.Is(err, windows.ERROR_SUCCESS) {
		panic(err)
	}

	memory := (*C.SharedMemory)(unsafe.Pointer(pointer))
	handle_video := func() {
		lock.Call(pointer)
		defer unlock.Call(pointer)

		block := memory.queues[C.Video0].array[memory.queues[C.Video0].order[0]]
		fmt.Printf("video buffer %d\n", block.size)

		for i := 0; i < C.QUEUE_SIZE-1; i++ {
			memory.queues[C.Video0].order[i] = memory.queues[C.Video0].order[i+1]
		}

		memory.queues[C.Video0].order[C.QUEUE_SIZE-1] = -1
	}

	handle_audio := func() {
		lock.Call(pointer)
		defer unlock.Call(pointer)

		block := memory.queues[C.Audio].array[memory.queues[C.Audio].order[0]]
		fmt.Printf("audio buffer %d\n", block.size)

		for i := 0; i < C.QUEUE_SIZE-1; i++ {
			memory.queues[C.Audio].order[i] = memory.queues[C.Audio].order[i+1]
		}

		memory.queues[C.Audio].order[C.QUEUE_SIZE-1] = -1
	}

	go func() {
		for {
			for peek(memory, video) {
				handle_video()
			}
			for peek(memory, audio) {
				handle_audio()
			}

			time.Sleep(time.Millisecond)
		}
	}()
	return (*SharedMemory)(unsafe.Pointer(pointer)),nil
}
