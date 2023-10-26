package win32

/*
#include <Windows.h>
*/
import "C"

func HighPriorityThread() {
	C.SetThreadPriority(C.GetCurrentThread(), C.THREAD_PRIORITY_HIGHEST)
}