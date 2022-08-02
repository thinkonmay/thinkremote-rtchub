// Package gst provides an easy API to create an appsink pipeline
package gst

/*
#cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0

#include "gst.h"

*/
import "C"
import (
	"fmt"
	"time"
	"unsafe"

	"github.com/pigeatgarlic/webrtc-proxy/listener"
	"github.com/pigeatgarlic/webrtc-proxy/util/config"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media"
)

func init() {
	go C.gstreamer_send_start_mainloop()
}

// Pipeline is a wrapper for a GStreamer Pipeline
type Pipeline struct {
	Pipeline  *C.GstElement
	sampchan  chan *media.Sample
	config    *config.ListenerConfig
	fun 	  listener.OnCloseFunc
}

var pipeline *Pipeline;

const (
	videoClockRate = 90000
	audioClockRate = 48000
	pcmClockRate   = 8000
)

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(config *config.ListenerConfig) *Pipeline {
	DIRECTX_PAD := "video/x-raw(memory:D3D11Memory)"
	QUEUE := "queue max-size-time=0 max-size-bytes=0 max-size-buffers=3" 
	pipelineStr := fmt.Sprintf("d3d11screencapturesrc blocksize=8192 show-cursor=true ! %s,framerate=60/1 ! %s ! d3d11convert ! %s,format=NV12 ! %s ! mfh264enc bitrate=5000 rc-mode=0 low-latency=true ref=1 quality-vs-speed=0 ! %s ! appsink name=appsink",DIRECTX_PAD,QUEUE,DIRECTX_PAD,QUEUE,QUEUE);
	// temp := fmt.Sprintf("videotestsrc ! video/x-raw,format=I420 ! x264enc speed-preset=ultrafast tune=zerolatency key-int-max=20 ! video/x-h264,stream-format=byte-stream ! appsink name=appsink")
	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))


	pipeline = &Pipeline{
		Pipeline:  C.gstreamer_send_create_pipeline(pipelineStrUnsafe),
		sampchan: make(chan *media.Sample),
		config: config,
	}

	return pipeline
}

//export goHandlePipelineBuffer
func goHandlePipelineBuffer(buffer unsafe.Pointer, bufferLen C.int, duration C.int) {
	c_byte := C.GoBytes(buffer, bufferLen)
	c_byteCopy := make([]byte, len(c_byte))
	copy(c_byteCopy, c_byte)
	
	sample := media.Sample{
		Data: c_byteCopy, 
		Duration: time.Duration(duration),
	}

	pipeline.sampchan <- &sample;	
	C.free(buffer)
}

func (p *Pipeline)ReadConfig() *config.ListenerConfig{
	return p.config;
}
func (p *Pipeline)ReadSample() *media.Sample{
	return <-p.sampchan;
}
func (p *Pipeline)ReadRTP() *rtp.Packet{
	block := make(chan *rtp.Packet)
	return <-block;
}




func (p *Pipeline) Open() {
	C.gstreamer_send_start_pipeline(p.Pipeline)
}
func (p *Pipeline)Close() {
	C.gstreamer_send_stop_pipeline(p.Pipeline)
}
func (p *Pipeline)OnClose(fun listener.OnCloseFunc) {
	p.fun = fun;
}
