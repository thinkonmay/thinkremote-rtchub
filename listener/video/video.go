// Package gst provides an easy API to create an appsink pipeline
package video

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay/h264"
	"github.com/thinkonmay/thinkremote-rtchub/listener/video/adaptive"
	"github.com/thinkonmay/thinkremote-rtchub/util/config"
)

// #cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0 gstreamer-video-1.0
// #cgo LDFLAGS: ${SRCDIR}/../cgo/lib/liblistener.a
// #include "webrtc_video.h"
import "C"

func init() {
	go C.start_video_mainloop()
}

// Pipeline is a wrapper for a GStreamer Pipeline
type Pipeline struct {
	closed     bool
	pipeline   unsafe.Pointer
	properties map[string]int

	pipelineStr string
	clockRate   float64


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

	packetizer rtppay.Packetizer
	codec      string

	adsContext *adaptive.AdaptiveContext

	restartCount int
}

var pipeline *Pipeline

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(pipelineStr string,
					Ads *config.DataChannel,
					Manual *config.DataChannel,
					) (*Pipeline,error) {

	pipeline = &Pipeline{
		closed: 	 false,	
		pipeline:    unsafe.Pointer(nil),
		pipelineStr: pipelineStr,
		codec:       webrtc.MimeTypeH264,

		clockRate:    90000,
		restartCount: 0,

		properties: make(map[string]int),
		adsContext: adaptive.NewAdsContext(Ads.Recv,
			func(bitrate int) { pipeline.SetProperty("bitrate", bitrate) }, 
			func() 			  { pipeline.SetProperty("reset", 0) },
		),
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
	fmt.Printf("starting video pipeline")
	pipeline.pipeline = C.create_video_pipeline(pipelineStrUnsafe, &err)
	err_str := ToGoString(err)
	if len(err_str) != 0 {
		return nil,fmt.Errorf("failed to create pipeline %s",err_str); 
	}


	go func() {
		for {
			data := <-Manual.Recv
			if pipeline.closed { return }
			var dat map[string]interface{}
			err := json.Unmarshal([]byte(data), &dat)
			if err != nil {
				fmt.Printf("%s", err.Error())
				continue
			}

			pipeline.SetProperty(dat["type"].(string), int(dat[dat["type"].(string)].(float64)))
		}
	}()

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

//export goHandlePipelineBufferVideo
func goHandlePipelineBufferVideo(buffer unsafe.Pointer, bufferLen C.int, duration C.int) {
	samples := uint32(time.Duration(duration).Seconds() * pipeline.clockRate)
	c_byte := C.GoBytes(buffer, bufferLen)
	pipeline.Multiplexer.srcBuf<-struct{buff *[]byte; samples int}{
		buff: &c_byte,
		samples: int(samples),
	}
}

func (p *Pipeline) GetCodec() string {
	return p.codec
}

func (p *Pipeline) SetProperty(name string, val int) error {
	fmt.Printf("%s change to %d\n", name, val)
	if p.pipeline == nil || p.properties == nil {
		return fmt.Errorf("attemping to set property while pipeline is not running, aborting");
	}

	switch name {
	case "bitrate":
		pipeline.properties["bitrate"] = val
		C.video_pipeline_set_bitrate(pipeline.pipeline, C.int(val))
	case "framerate":
		pipeline.properties["framerate"] = val
		C.video_pipeline_set_framerate(pipeline.pipeline, C.int(val))
	case "reset":
		C.force_gen_idr_frame_video_pipeline(pipeline.pipeline)
	default:
		return fmt.Errorf("unknown prop")
	}
	return nil
}

//export handleVideoStopOrError
func handleVideoStopOrError() {
	pipeline.Close()
	for key, val := range pipeline.properties {
		pipeline.SetProperty(key, val)
	}
	pipeline.Open()

	pipeline.restartCount++
}

func (p *Pipeline) Open() {
	p.packetizer = h264.NewH264Payloader()
	C.start_video_pipeline(pipeline.pipeline)
}
func (p *Pipeline) Close() {
	p.packetizer = nil
	C.stop_video_pipeline(p.pipeline)
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