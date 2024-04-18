package proxy

/*
#include <string.h>
*/
import "C"
import (
	"errors"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func memcpy(to,from unsafe.Pointer, size int) {
	C.memcpy(to, from, C.ulonglong(size))
}

func ObtainSharedMemory(token string) (*SharedMemory, error) {
	mod, err := syscall.LoadDLL("libparent.dll")
	if err != nil {
		return nil, err
	}
	deinit, err := mod.FindProc("deinit_shared_memory")
	if err != nil {
		return nil, err
	}
	deinit.Call() // actually we don't allocate new memory, 
	obtain, err := mod.FindProc("obtain_shared_memory")
	if err != nil {
		return nil, err
	}

	buffer := []byte(token)
	pointer, _, err := obtain.Call(
		uintptr(unsafe.Pointer(&buffer[0])),
	)
	if !errors.Is(err, windows.ERROR_SUCCESS) {
		return nil,err
	}

	return (*SharedMemory)(unsafe.Pointer(pointer)), nil
}
