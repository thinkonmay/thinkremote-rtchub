// Package gst provides an easy API to create an appsink pipeline
package audio

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"github.com/thinkonmay/thinkremote-rtchub/listener/multiplexer"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay/opus"
)

// #cgo LDFLAGS: ${SRCDIR}/../../cgo/lib/libshared.a
// #cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0
// #include "webrtc_audio.h"
import "C"

func init() {
	go C.start_audio_mainloop()
}

// Pipeline is a wrapper for a GStreamer Pipeline
type Pipeline struct {
	closed     bool
	pipeline    unsafe.Pointer
	pipelineStr string

	clockRate float64

	codec      string

	restartCount int

	Multiplexer *multiplexer.Multiplexer
}

var pipeline *Pipeline

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(pipelinestr string) (*Pipeline,error) {
	pipeline = &Pipeline{
		pipeline:     unsafe.Pointer(nil),
		pipelineStr:  pipelinestr,
		restartCount: 0,
		clockRate:    48000,
		codec:        webrtc.MimeTypeOpus,

		Multiplexer: multiplexer.NewMultiplexer("audio",func() rtppay.Packetizer {
			return opus.NewOpusPayloader()
		}),
	}

	pipelineStrUnsafe := C.CString(pipeline.pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	var err unsafe.Pointer
	fmt.Printf("starting audio pipeline %s\n",pipelinestr)
	pipeline.pipeline = C.create_audio_pipeline(pipelineStrUnsafe, &err)
	err_str := ToGoString(err)
	if len(err_str) != 0 {
		return nil,fmt.Errorf("fail to create pipeline %s",err_str);
	}


	return pipeline,nil
}

//export goHandlePipelineBufferAudio
func goHandlePipelineBufferAudio(buffer unsafe.Pointer, bufferLen C.int, duration C.int) {
	samples := uint32(time.Duration(duration).Seconds() * pipeline.clockRate)
	pipeline.Multiplexer.Send(buffer, uint32(bufferLen), uint32(samples) )
}

//export handleAudioStopOrError
func handleAudioStopOrError() {
	pipeline.Close()
	pipeline.Open()

	pipeline.restartCount++
}

func (p *Pipeline) GetCodec() string {
	return p.codec
}

func (p *Pipeline) Open() {
	fmt.Println("starting audio pipeline")
	C.start_audio_pipeline(pipeline.pipeline)
}


func (p *Pipeline) Close() {
	fmt.Println("stoping audio pipeline")
	C.stop_audio_pipeline(p.pipeline)
}

func (p *Pipeline) SetProperty(name string, val int) error {
	return fmt.Errorf("unknown prop")
}

func ToGoString(str unsafe.Pointer) string {
	if str == nil {
		return ""
	}
	return string(C.GoBytes(str, C.int(C.string_get_length(str))))
}




func (p *Pipeline) RegisterRTPHandler(id string, fun func(pkt *rtp.Packet)) {
	p.Multiplexer.RegisterRTPHandler(id,fun)
}

func (p *Pipeline) DeregisterRTPHandler(id string) {
	p.Multiplexer.DeregisterRTPHandler(id)
}