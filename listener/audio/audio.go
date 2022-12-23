// Package gst provides an easy API to create an appsink pipeline
package audio

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/OnePlay-Internet/webrtc-proxy/rtppay"
	"github.com/OnePlay-Internet/webrtc-proxy/rtppay/opus"
	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	gsttest "github.com/OnePlay-Internet/webrtc-proxy/util/test"
	"github.com/OnePlay-Internet/webrtc-proxy/util/tool"
	"github.com/pion/rtp"
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
	pipeline unsafe.Pointer
	pipelineStr string

	soundcard *tool.Soundcard
	clockRate float64

	rtpchan chan *rtp.Packet
	config  *config.ListenerConfig

	packetizer rtppay.Packetizer

	restartCount int

	lastPacketTimestamp time.Time
}

var pipeline *Pipeline

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(config *config.ListenerConfig) *Pipeline {
	pipeline = &Pipeline{
		pipeline:  unsafe.Pointer(nil),
		rtpchan:   make(chan *rtp.Packet),
		config:    config,
		pipelineStr : "fakesrc ! appsink name=appsink",
		restartCount: 0,
		clockRate: gsttest.AudioClockRate,

		soundcard: &tool.Soundcard{
			DeviceID: "none",
		},

		packetizer: opus.NewOpusPayloader(),
	}
	return pipeline
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

func (p *Pipeline) GetSourceName() (string) {
	return p.soundcard.DeviceID;
}
func (p *Pipeline) SetProperty(name string,val int) error {
	return fmt.Errorf("unknown prop");
}



func (p *Pipeline) SetSource(source interface{}) error {
	if source.(*tool.Soundcard).DeviceID != "none" {
		pipelineStr := gsttest.GstTestAudio(source.(*tool.Soundcard))
		if pipelineStr == "" {
			return fmt.Errorf("unable to create encode pipeline with device")
		}
		p.pipelineStr = pipelineStr
	}

	pipelineStrUnsafe := C.CString(p.pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	var err unsafe.Pointer
	Pipeline := C.create_audio_pipeline(pipelineStrUnsafe, &err)
	if len(tool.ToGoString(err)) != 0 {
		C.stop_audio_pipeline(Pipeline)
		return fmt.Errorf("%s", tool.ToGoString(err))
	}


	p.pipeline = Pipeline
	return nil
}

//export handleAudioStopOrError
func handleAudioStopOrError() {
	pipeline.Close()
	pipeline.SetSource(pipeline.soundcard)
	pipeline.Open()
	pipeline.restartCount++
}

func (p *Pipeline) Open() {
	fmt.Printf("starting audio pipeline: %s\n", p.pipelineStr)
	C.start_audio_pipeline(pipeline.pipeline)
}
func (p *Pipeline) GetConfig() *config.ListenerConfig {
	return p.config
}
func (p *Pipeline) ReadRTP() *rtp.Packet {
	return <-p.rtpchan
}

func (p *Pipeline) Close() {
	C.stop_audio_pipeline(p.pipeline)
}
