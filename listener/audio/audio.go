// Package gst provides an easy API to create an appsink pipeline
package audio

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/pion/rtp"
	"github.com/thinkonmay/thinkremote-rtchub/rtppay"
	"github.com/thinkonmay/thinkremote-rtchub/rtppay/opus"
	"github.com/thinkonmay/thinkremote-rtchub/util/config"
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
	pipeline    unsafe.Pointer
	pipelineStr string

	clockRate float64

	rtpchan chan *rtp.Packet
	packetizer rtppay.Packetizer
	restartCount int
}

var pipeline *Pipeline

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(config *config.ListenerConfig) (*Pipeline,error) {
	pipeline = &Pipeline{
		pipeline:     unsafe.Pointer(nil),
		rtpchan:      make(chan *rtp.Packet),
		pipelineStr:  "fakesrc ! appsink name=appsink",
		restartCount: 0,
		clockRate:    gsttest.AudioClockRate,

		packetizer: opus.NewOpusPayloader(),
	}

	pipelineStrUnsafe := C.CString(pipeline.pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	var err unsafe.Pointer
	Pipeline := C.create_audio_pipeline(pipelineStrUnsafe, &err)
	if len(tool.ToGoString(err)) != 0 {
		C.stop_audio_pipeline(Pipeline)
		return nil,fmt.Errorf("%s", tool.ToGoString(err))
	}

	pipeline.pipeline = Pipeline
	return pipeline,nil
}

//export goHandlePipelineBufferAudio
func goHandlePipelineBufferAudio(buffer unsafe.Pointer, bufferLen C.int, duration C.int) {
	samples := uint32(time.Duration(duration).Seconds() * pipeline.clockRate)
	c_byte := C.GoBytes(buffer, bufferLen)
	packets := pipeline.packetizer.Packetize(c_byte, samples)

	for _, packet := range packets {
		pipeline.rtpchan <- packet
	}
}



//export handleAudioStopOrError
func handleAudioStopOrError() {
	pipeline.Close()
	pipeline.Open()

	pipeline.restartCount++
}

func (p *Pipeline) Open() {
	fmt.Println("starting audio pipeline")
	C.start_audio_pipeline(pipeline.pipeline)
}

func (p *Pipeline) ReadRTP() *rtp.Packet {
	return <-p.rtpchan
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
