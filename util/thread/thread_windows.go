package thread

/*
#include <Windows.h>
typedef enum _D3DKMT_SCHEDULINGPRIORITYCLASS {
	D3DKMT_SCHEDULINGPRIORITYCLASS_IDLE,
	D3DKMT_SCHEDULINGPRIORITYCLASS_BELOW_NORMAL,
	D3DKMT_SCHEDULINGPRIORITYCLASS_NORMAL,
	D3DKMT_SCHEDULINGPRIORITYCLASS_ABOVE_NORMAL,
	D3DKMT_SCHEDULINGPRIORITYCLASS_HIGH,
	D3DKMT_SCHEDULINGPRIORITYCLASS_REALTIME
} D3DKMT_SCHEDULINGPRIORITYCLASS;

typedef UINT D3DKMT_HANDLE;

typedef struct _D3DKMT_OPENADAPTERFROMLUID {
	LUID AdapterLuid;
	D3DKMT_HANDLE hAdapter;
} D3DKMT_OPENADAPTERFROMLUID;

typedef struct _D3DKMT_QUERYADAPTERINFO {
	D3DKMT_HANDLE hAdapter;
	UINT Type;
	VOID *pPrivateDriverData;
	UINT PrivateDriverDataSize;
} D3DKMT_QUERYADAPTERINFO;

typedef struct _D3DKMT_CLOSEADAPTER {
	D3DKMT_HANDLE hAdapter;
} D3DKMT_CLOSEADAPTER;

typedef NTSTATUS(WINAPI *PD3DKMTSetProcessSchedulingPriorityClass)(HANDLE, D3DKMT_SCHEDULINGPRIORITYCLASS);
typedef NTSTATUS(WINAPI *PD3DKMTOpenAdapterFromLuid)(D3DKMT_OPENADAPTERFROMLUID *);
typedef NTSTATUS(WINAPI *PD3DKMTQueryAdapterInfo)(D3DKMT_QUERYADAPTERINFO *);
typedef NTSTATUS(WINAPI *PD3DKMTCloseAdapter)(D3DKMT_CLOSEADAPTER *);

int SetGPURealtimePriority() {
    HMODULE gdi32 = GetModuleHandleA("GDI32");

	PD3DKMTSetProcessSchedulingPriorityClass d3dkmt_set_process_priority =
		(PD3DKMTSetProcessSchedulingPriorityClass) GetProcAddress(gdi32, "D3DKMTSetProcessSchedulingPriorityClass");

	if (!d3dkmt_set_process_priority)
		return 0;

	D3DKMT_SCHEDULINGPRIORITYCLASS priority = D3DKMT_SCHEDULINGPRIORITYCLASS_REALTIME;

	// int hags_enabled = check_hags(adapter_desc.AdapterLuid);
	// if (adapter_desc.VendorId == 0x10DE) {
	// As of 2023.07, NVIDIA driver has unfixed bug(s) where "realtime" can cause unrecoverable encoding freeze or outright driver crash
	// This issue happens more frequently with HAGS, in DX12 games or when VRAM is filled close to max capacity
	// Track OBS to see if they find better workaround or NVIDIA fixes it on their end, they seem to be in communication
	// if (hags_enabled && !config::video.nv_realtime_hags) priority = D3DKMT_SCHEDULINGPRIORITYCLASS_HIGH;
	// }
	// BOOST_LOG(info) << "Active GPU has HAGS " << (hags_enabled ? "enabled" : "disabled");
	// BOOST_LOG(info) << "Using " << (priority == D3DKMT_SCHEDULINGPRIORITYCLASS_HIGH ? "high" : "realtime") << " GPU priority";

	if (FAILED(d3dkmt_set_process_priority(GetCurrentProcess(), priority)))
		return 0;

	return 1;
}

*/
import "C"


func init() {
	resulta := C.SetPriorityClass(C.GetCurrentProcess(), C.REALTIME_PRIORITY_CLASS)
	resultb := C.SetGPURealtimePriority()
	if resulta == 0 || resultb == 0 {
		fmt.Printf("failed to set realtime priority\n")
	} else {
		fmt.Printf("set realtime priority\n")
	}
}



func HighPriorityThread() {
	C.SetThreadPriority(C.GetCurrentThread(), C.THREAD_PRIORITY_HIGHEST)
}