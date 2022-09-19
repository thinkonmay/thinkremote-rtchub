package sink

import (
	"fmt"
	"unsafe"

	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	"github.com/pion/rtp"
)

// #cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0
// #cgo LDFLAGS: ${SRCDIR}/../../lib/libshared.a
// #include "sink.h"
import "C"

func init() {
	go C.start_sink_mainloop()
}

// Pipeline is a wrapper for a GStreamer Pipeline
type VideoSink struct {
	Pipeline *C.GstElement
	config *config.BroadcasterConfig
}
// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(config *config.BroadcasterConfig) (*VideoSink,error) {

	pipelineStr := "appsrc format=time is-live=true do-timestamp=true name=src ! application/x-rtp"
	pipelineStr += " ! queue ! rtph264depay ! queue ! decodebin ! queue ! autovideosink"

	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))
	ret := &VideoSink {
		Pipeline: C.create_sink_pipeline(pipelineStrUnsafe),
		config: config,
	}
	C.start_sink_pipeline(ret.Pipeline)
	return ret,nil;
}



// Push pushes a buffer on the appsrc of the GStreamer Pipeline
func (p *VideoSink) Push(buffer []byte) {
	b := C.CBytes(buffer)
	defer C.free(b)
	C.push_sink_buffer(p.Pipeline, b, C.int(len(buffer)))
}





func (sink *VideoSink) Write(packet *rtp.Packet) {
	buf,err := packet.Marshal()
	if err != nil {
		fmt.Printf("%s\n",err.Error());
		return;		
	}
	sink.Push(buf)
}

func (p *VideoSink) Close() {
	C.stop_sink_pipeline(p.Pipeline)
}

