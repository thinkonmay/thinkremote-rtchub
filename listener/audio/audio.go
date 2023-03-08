// Package gst provides an easy API to create an appsink pipeline
package audio

import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay/opus"
)

// #cgo LDFLAGS: ${SRCDIR}/../cgo/lib/liblistener.a
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

	rtpchan    chan *rtp.Packet
	packetizer rtppay.Packetizer
	codec      string

	restartCount int

	Multiplexer struct {
		srcPkt    		chan *rtp.Packet
		srcBuf    		chan struct {
			buff    *[]byte
			samples int
		}

		mutex 		*sync.Mutex
		handler 	map[string]struct {
			closed			bool
			sink    		chan *rtp.Packet
			handler         func(*rtp.Packet)
		}	
	}
}

var pipeline *Pipeline

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(pipelinestr string) (*Pipeline,error) {
	pipeline = &Pipeline{
		pipeline:     unsafe.Pointer(nil),
		rtpchan:      make(chan *rtp.Packet),
		pipelineStr:  pipelinestr,
		restartCount: 0,
		clockRate:    48000,
		codec:        webrtc.MimeTypeOpus,

		packetizer: opus.NewOpusPayloader(),
		Multiplexer: struct{
			srcPkt chan *rtp.Packet; 
			srcBuf chan struct{buff *[]byte; samples int}; 
			mutex 	*sync.Mutex;
			handler map[string]struct{closed bool; sink chan *rtp.Packet; handler func(*rtp.Packet)};
		}{
			srcPkt: make(chan *rtp.Packet,50),
			srcBuf: make(chan struct{buff *[]byte; samples int},50),
			mutex:  &sync.Mutex{},
			handler: map[string]struct{closed bool; sink chan *rtp.Packet; handler func(*rtp.Packet)}{},
		},
	}

	pipelineStrUnsafe := C.CString(pipeline.pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	var err unsafe.Pointer
	fmt.Printf("starting audio pipeline")
	pipeline.pipeline = C.create_audio_pipeline(pipelineStrUnsafe, &err)
	err_str := ToGoString(err)
	if len(err_str) != 0 {
		return nil,fmt.Errorf("fail to create pipeline %s",err_str);
	}


	go func() {
		for {
			src_buffer := <- pipeline.Multiplexer.srcBuf
			if pipeline.closed { return }
			packets := pipeline.packetizer.Packetize(*src_buffer.buff, uint32(src_buffer.samples))
			for _, packet := range packets {
				pipeline.Multiplexer.srcPkt <- packet
			}
		}
	}()

	go func() {
		for {
			src_pkt := <- pipeline.Multiplexer.srcPkt
			if pipeline.closed { return }
			pipeline.Multiplexer.mutex.Lock()
			for _,v := range pipeline.Multiplexer.handler {
				if v.closed { continue; }
				v.sink <- src_pkt.Clone()
			}
			pipeline.Multiplexer.mutex.Unlock()
		}
	}()
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

func (p *Pipeline) GetCodec() string {
	return p.codec
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


func (p *Pipeline) RegisterRTPHandler(id string, fun func(pkt *rtp.Packet)) {
	if p.Multiplexer.handler == nil {
		fmt.Println("Try to register RTP handler while pipeline not ready")
		return
	}


	handler := struct{
		closed bool; 
		sink chan *rtp.Packet; 
		handler func(*rtp.Packet);
	}{
		closed: false,
		sink: make(chan *rtp.Packet,50),
		handler: fun,
	}

	go func ()  {
		for {
			pkt := <-handler.sink
			if handler.closed { return }
			handler.handler(pkt);
		}
	}()

	p.Multiplexer.mutex.Lock()
	p.Multiplexer.handler[id] = handler;
	p.Multiplexer.mutex.Unlock()
}

func (p *Pipeline) DeregisterRTPHandler(id string) {
	p.Multiplexer.mutex.Lock()
	defer p.Multiplexer.mutex.Unlock()

	handler := p.Multiplexer.handler[id];
	handler.closed = true

	delete(p.Multiplexer.handler,id)
}