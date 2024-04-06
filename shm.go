package proxy

/*
#include "smemory.h"
#include <string.h>
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
	Max        = C.QueueMax

	Idr        = C.Idr
)

func (mem *SharedMemory) GetQueue(id int) *Queue {
	return (*Queue)(&mem.queues[id])
}

func (memory *Queue) Raise(event_id, value int) {
	memory.events[event_id].value_number = C.int(value)
	memory.events[event_id].read = 0
}
func (queue *Queue) CurrentIndex() int {
	return int(queue.index)
}
func (queue *Queue) Copy(in []byte, index int) int {
	real_index := index % int(C.QUEUE_SIZE)
	block := queue.array[real_index]
	C.memcpy(unsafe.Pointer(&in[0]), unsafe.Pointer(&block.data[0]), C.ulonglong(block.size))
	return int(block.size)
}
