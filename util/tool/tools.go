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
    MonitorHandle int   `json:"handle"`
    MonitorName string	`json:"name"`
    DeviceName string	`json:"device"`
    Adapter string 		`json:"adapter"`
	Width int 			`json:"width"`
	Height int 			`json:"height"`
	IsPrimary bool 		`json:"isPrimary"`
};

type Soundcard struct {
    DeviceID string 	`json:"id"`	
    Name string 		`json:"name"`	
	Api string			`json:"api"`	

	IsDefault bool 		`json:"isDefault"`
	IsLoopback bool 	`json:"isLoopback"`
};

type MediaDevice struct {
    Monitors []Monitor       `json:"monitors"`
    Soundcards []Soundcard 	 `json:"soundcards"`
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
		width    := 			C.get_monitor_width(query,count_monitor);
		height   := 			C.get_monitor_height(query,count_monitor);
		prim     := 			C.monitor_is_primary(query,count_monitor);

		result.Monitors = append(result.Monitors, Monitor{
			MonitorHandle: int(mhandle),
			MonitorName: string(C.GoBytes(monitor_name,C.string_get_length(monitor_name))),
			Adapter: string(C.GoBytes(adapter,C.string_get_length(adapter))),
			DeviceName: string(C.GoBytes(device_name,C.string_get_length(device_name))),
			Width: int(width),
			Height: int(height),
			IsPrimary: (prim == 1),	
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
		api := 				C.get_soundcard_api(query,count_soundcard);
		loopback:= 		    C.soundcard_is_loopback(query,count_soundcard);
		defaul:= 		    C.soundcard_is_default(query,count_soundcard);

		result.Soundcards = append(result.Soundcards,Soundcard{
			Name: string(C.GoBytes(name,C.string_get_length(name))),
			DeviceID: string(C.GoBytes(device_id,C.string_get_length(device_id))),
			Api: string(C.GoBytes(api,C.string_get_length(api))),
			IsDefault: (defaul == 1),
			IsLoopback: (loopback == 1),

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