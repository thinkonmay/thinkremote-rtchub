package proxy

/*
#include <string.h>
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
	"unsafe"
)

type SharedMemory C.SharedMemory

func (memory *SharedMemory) Raise(event_id, value int) {
	memory.events[event_id].value_number = C.int(value)
	memory.events[event_id].read = 0
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

func (memory *SharedMemory) Peek(media int) bool {
	return memory.queues[media].order[0] != -1
}
func (memory *SharedMemory) Copy(in []byte, media int) int {
	memory.Lock()
	defer memory.Unlock()

	block := memory.queues[media].array[memory.queues[media].order[0]]
	C.memcpy(unsafe.Pointer(&in[0]), unsafe.Pointer(&block.data[0]), C.ulonglong(block.size))

	memory.queues[media].order[C.QUEUE_SIZE-1] = -1
	for i := 0; i < C.QUEUE_SIZE-1; i++ {
		memory.queues[media].order[i] = memory.queues[media].order[i+1]
	}

	return int(block.size)
}
