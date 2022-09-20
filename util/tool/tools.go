package tool
import (
	"unsafe"
)

// #cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0
// #cgo LDFLAGS: ${SRCDIR}/../../lib/libshared.a
// #include "util.h"
import "C"


type DeviceQuery unsafe.Pointer

type Monitor struct {
    MonitorHandle int
    MonitorName string
    DeviceName string;
    Adapter string;
};

type Soundcard struct {
    DeviceID string;
    Name string;
};

type MediaDevice struct {
    Monitors []Monitor;
    Soundcards []Soundcard;
};


func GetDevice() *MediaDevice{
	result := &MediaDevice{};
	query := C.query_media_device()

	count_soundcard := C.int(0);
	count_monitor := C.int(0);
	for {
		active := C.monitor_is_active(query,count_monitor);
		if active == 0 {
			break;
		}
		mhandle := 				C.get_monitor_handle(query,count_monitor);
		monitor_name := 		C.get_monitor_name(query,count_monitor);
		adapter := 				C.get_monitor_adapter(query,count_monitor);
		device_name := 			C.get_monitor_device_name(query,count_monitor);

		result.Monitors = append(result.Monitors, Monitor{
			MonitorHandle: int(mhandle),
			MonitorName: string(C.GoBytes(monitor_name,C.string_get_length(monitor_name))),
			Adapter: string(C.GoBytes(adapter,C.string_get_length(adapter))),
			DeviceName: string(C.GoBytes(device_name,C.string_get_length(device_name))),
		})
		count_monitor++;
	}

	for {
		active := C.soundcard_is_active(query,count_soundcard);
		if active == 0 {
			break;
		}
		name:= 				C.get_soundcard_name(query,count_soundcard);
		device_id:= 		C.get_soundcard_device_id(query,count_soundcard);

		result.Soundcards = append(result.Soundcards,Soundcard{
			Name: string(C.GoBytes(name,C.string_get_length(name))),
			DeviceID: string(C.GoBytes(device_id,C.string_get_length(device_id))),
		})
		count_soundcard++;
	}

	return result;
}



func ToGoString(str unsafe.Pointer) string {
	if str == nil {
		return ""
	}
	return string(C.GoBytes(str,C.int(C.string_get_length(str))));
}