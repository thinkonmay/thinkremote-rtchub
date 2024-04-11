package proxy

import (
	"unsafe"

	"github.com/ebitengine/purego"
)

func ObtainSharedMemory(token string) (*SharedMemory, error) {
	libc, err := purego.Dlopen("./libparent.so", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return nil, err
	}

	var deinit func()
	purego.RegisterLibFunc(&deinit, libc, "deinit_shared_memory")
	if err != nil {
		return nil, err
	}
	deinit() // actually we don't allocate new memory,

	var allocate func(unsafe.Pointer) unsafe.Pointer
	purego.RegisterLibFunc(&allocate, libc, "obtain_shared_memory")
	pointer := allocate(unsafe.Pointer(&[]byte(token)[0]))

	return (*SharedMemory)(unsafe.Pointer(pointer)), nil
}
