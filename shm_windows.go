package proxy

import (
	"errors"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	lock   *syscall.Proc
	unlock *syscall.Proc
)

func (memory *SharedMemory) Lock() {
	lock.Call(uintptr(unsafe.Pointer(memory)))
}
func (memory *SharedMemory) Unlock() {
	unlock.Call(uintptr(unsafe.Pointer(memory)))
}

func ObtainSharedMemory(token string) (*SharedMemory, error) {
	mod, err := syscall.LoadDLL("libparent.dll")
	if err != nil {
		return nil, err
	}
	obtain, err := mod.FindProc("obtain_shared_memory")
	if err != nil {
		return nil, err
	}
	lock, err = mod.FindProc("lock_shared_memory")
	if err != nil {
		return nil, err
	}
	unlock, err = mod.FindProc("unlock_shared_memory")
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

	return (*SharedMemory)(unsafe.Pointer(pointer)), nil
}
