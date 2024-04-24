package proxy

/*
#include "smemory.h"
*/
import "C"
import (
	"unsafe"
)

type SharedMemory C.SharedMemory
type Queue C.Queue

const (
	Audio      = C.Audio
	Microphone = C.Microphone
	Input      = C.Input
	Max        = C.QueueMax
	Idr        = C.Idr
	Framerate  = C.Framerate
	Bitrate    = C.Bitrate
	Pointer    = C.Pointer
)

func (mem *SharedMemory) GetQueue(id int) *Queue {
	return (*Queue)(&mem.queues[id])
}

func (memory *Queue) Raise(event_id, value int) {
	memory.events[event_id].value_number = C.int(value)
	memory.events[event_id].read = 0
}
func (queue *Queue) GetDisplay() string {
	return C.GoString(&queue.metadata.display[0])
}
func (queue *Queue) CurrentIndex() int {
	return int(queue.index)
}
func (queue *Queue) Copy(in []byte, index int) (size int, duration int64) {
	real_index := index % int(C.QUEUE_SIZE)
	block := &queue.array[real_index]
	memcpy(unsafe.Pointer(&in[0]), unsafe.Pointer(&block.data[0]), int(block.size))
	return int(block.size),int64(block.metadata.duration)
}

func (queue *Queue) Write(in []byte, size int) {
	new_idx := queue.index + 1
	block := &queue.array[new_idx%C.QUEUE_SIZE]
	memcpy(unsafe.Pointer(&block.data[0]), unsafe.Pointer(&in[0]), int(block.size))
	queue.index = new_idx
}
