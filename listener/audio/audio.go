package audio

import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"github.com/thinkonmay/thinkremote-rtchub/listener/multiplexer"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay/opus"
	"github.com/thinkonmay/thinkremote-rtchub/util/sunshine"
	"github.com/thinkonmay/thinkremote-rtchub/util/win32"
)

import "C"


type AudioPipelineC unsafe.Pointer
type AudioPipeline struct {
	closed      bool
	pipeline    unsafe.Pointer
	mut        *sync.Mutex

	clockRate float64

	codec string
	Multiplexer *multiplexer.Multiplexer
}


// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline() (*AudioPipeline, error) {
	pipeline := &AudioPipeline{
		closed:      false,
		clockRate:   48000,
		codec:       webrtc.MimeTypeOpus,

		Multiplexer: multiplexer.NewMultiplexer("audio", func() rtppay.Packetizer {
			return opus.NewOpusPayloader()
		}),
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


func (pipeline *AudioPipeline) reset() {
	pipeline.mut.Lock()
	defer pipeline.mut.Unlock()
	if pipeline.pipeline != nil {
		sunshine.RaiseEvent(pipeline.pipeline,sunshine.STOP,0)
		sunshine.WaitEvent(pipeline.pipeline,sunshine.STOP)
	}

	pipeline.pipeline =  sunshine.StartQueue(sunshine.OPUS);
}

func (p *AudioPipeline) GetCodec() string {
	return p.codec
}

func (p *AudioPipeline) Open() {
	fmt.Println("starting audio pipeline")
}

func (p *AudioPipeline) Close() {
	fmt.Println("stoping audio pipeline")
}

func (p *AudioPipeline) SetPropertyS(name string, val string) error {
	return fmt.Errorf("unknown prop")
}
func (p *AudioPipeline) SetProperty(name string, val int) error {
	return fmt.Errorf("unknown prop")
}


func (p *AudioPipeline) RegisterRTPHandler(id string, fun func(pkt *rtp.Packet)) {
	p.Multiplexer.RegisterRTPHandler(id, fun)
}

func (p *AudioPipeline) DeregisterRTPHandler(id string) {
	p.Multiplexer.DeregisterRTPHandler(id)
}
