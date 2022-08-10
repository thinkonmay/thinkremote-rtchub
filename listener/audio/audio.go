// Package gst provides an easy API to create an appsink pipeline
package audio

/*
#cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0

#include "audio.h"

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
	go C.gstreamer_audio_start_mainloop()
}

// Pipeline is a wrapper for a GStreamer Pipeline
type Pipeline struct {
	Pipeline *C.GstElement
	sampchan chan *media.Sample
	config   *config.ListenerConfig
	fun      listener.OnCloseFunc
}

var pipeline *Pipeline

const (
	videoClockRate = 90000
	audioClockRate = 48000
	pcmClockRate   = 8000
)

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(config *config.ListenerConfig) *Pipeline {
	dev := C.set_media_device();
	bytes := C.GoBytes(dev,C.string_get_length(dev));
	fmt.Printf("audio device id: %s\n",string(bytes));

	QUEUE := "queue max-size-time=0 max-size-bytes=0 max-size-buffers=3"
	pipelineStr := fmt.Sprintf("wasapisrc name=source loopback=true ! %s ! audioconvert ! %s ! audioresample ! %s ! opusenc ! %s ! appsink name=appsink", QUEUE, QUEUE, QUEUE, QUEUE)
	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	pipeline = &Pipeline{
		Pipeline: C.gstreamer_audio_create_pipeline(pipelineStrUnsafe,dev),
		sampchan: make(chan *media.Sample),
		config:   config,
	}

	return pipeline
}

//export goHandlePipelineBufferAudio
func goHandlePipelineBufferAudio(buffer unsafe.Pointer, bufferLen C.int, duration C.int) {
	defer C.free(buffer)

	c_byte := C.GoBytes(buffer, bufferLen)
	sample := media.Sample{
		Data:     c_byte,
		Duration: time.Duration(duration),
	}
	pipeline.sampchan <- &sample
}

func (p *Pipeline) ReadConfig() *config.ListenerConfig {
	return p.config
}
func (p *Pipeline) ReadSample() *media.Sample {
	return <-p.sampchan
}
func (p *Pipeline) ReadRTP() *rtp.Packet {
	block := make(chan *rtp.Packet)
	return <-block
}

func (p *Pipeline) Open() {
	C.gstreamer_audio_start_pipeline(p.Pipeline)
}
func (p *Pipeline) Close() {
	C.gstreamer_audio_stop_pipeline(p.Pipeline)
}
func (p *Pipeline) OnClose(fun listener.OnCloseFunc) {
	p.fun = fun
}
