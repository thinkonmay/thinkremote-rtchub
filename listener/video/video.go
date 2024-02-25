package video

import "C"
import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"github.com/thinkonmay/thinkremote-rtchub/listener"
	"github.com/thinkonmay/thinkremote-rtchub/listener/multiplexer"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay/h264"
	"github.com/thinkonmay/thinkremote-rtchub/util/win32"
	"github.com/thinkonmay/thinkremote-rtchub/util/sunshine"
)


type VideoPipelineC unsafe.Pointer
type VideoPipeline struct {
	closed     bool
	pipeline   unsafe.Pointer
	mut        *sync.Mutex
	properties map[string]int
	sproperties map[string]string

	clockRate   float64

	codec string
	Multiplexer *multiplexer.Multiplexer
}

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(display string) ( listener.Listener,
	                                  error) {

	pipeline := &VideoPipeline{
		closed:      false,
		pipeline:    nil,
		mut: 		 &sync.Mutex{},
		codec:       webrtc.MimeTypeH264,

		clockRate: 90000,

		properties: map[string]int{
			"codec": int(sunshine.H264),
			"bitrate": 6000,
		},
		sproperties: map[string]string{
			"display": display,
		},
		Multiplexer: multiplexer.NewMultiplexer("video", h264.NewH264Payloader() ),
	}



	pipeline.reset()
	go func() { win32.HighPriorityThread()
		buffer := make([]byte, 256*1024) //256kB
		timestamp := time.Now().UnixNano()

		for {
			pipeline.mut.Lock()
			size := sunshine.PopFromQueue(pipeline.pipeline,
                unsafe.Pointer(&buffer[0]))
			pipeline.mut.Unlock()
			if size == 0 {
				continue
			}

			diff := time.Now().UnixNano() - timestamp
			pipeline.Multiplexer.Send(buffer[:size], uint32(time.Duration(diff).Seconds() * pipeline.clockRate))
			timestamp = timestamp + diff
		}
	}()
	return pipeline, nil
}

func (p *VideoPipeline) GetCodec() string {
	return p.codec
}

func (pipeline *VideoPipeline) reset() {
	pipeline.mut.Lock()
	defer pipeline.mut.Unlock()
	if pipeline.pipeline != nil {
		sunshine.RaiseEvent(pipeline.pipeline,sunshine.STOP,0)
		sunshine.WaitEvent(pipeline.pipeline,sunshine.STOP)
	}

	pipeline.pipeline =  sunshine.StartQueue(pipeline.properties["codec"]);
	sunshine.RaiseEventS(pipeline.pipeline,sunshine.CHANGE_DISPLAY,pipeline.sproperties["display"])
}

func (pipeline *VideoPipeline) SetPropertyS(name string, val string) error {
	fmt.Printf("%s change to %s\n", name, val)
	switch name {
	case "display":
		pipeline.sproperties["display"] = val
		sunshine.RaiseEventS(pipeline.pipeline,sunshine.CHANGE_DISPLAY,pipeline.sproperties["display"])
	}
	return nil
}
func (pipeline *VideoPipeline) SetProperty(name string, val int) error {
	fmt.Printf("%s change to %d\n", name, val)
	switch name {
	case "bitrate":
		pipeline.properties["bitrate"] = val
		sunshine.RaiseEvent(pipeline.pipeline,sunshine.CHANGE_BITRATE,pipeline.properties["bitrate"])
	case "framerate":
		pipeline.properties["framerate"] = val
		sunshine.RaiseEvent(pipeline.pipeline,sunshine.CHANGE_FRAMERATE,pipeline.properties["framerate"])
	case "pointer":
		pipeline.properties["pointer"] = val
		sunshine.RaiseEvent(pipeline.pipeline,sunshine.POINTER_VISIBLE,pipeline.properties["pointer"])
	case "codec":
		pipeline.properties["codec"] = val
		pipeline.reset()
	case "reset":
		sunshine.RaiseEvent(pipeline.pipeline,sunshine.IDR_FRAME,0)
	default:
		return fmt.Errorf("unknown prop")
	}
	return nil
}

func (p *VideoPipeline) Open() {
	fmt.Println("starting video pipeline")
}
func (p *VideoPipeline) Close() {
	fmt.Println("stopping video pipeline")
}

func (p *VideoPipeline) RegisterRTPHandler(id string, fun func(pkt *rtp.Packet)) {
	p.Multiplexer.RegisterRTPHandler(id, fun)
}

func (p *VideoPipeline) DeregisterRTPHandler(id string) {
	p.Multiplexer.DeregisterRTPHandler(id)
}
