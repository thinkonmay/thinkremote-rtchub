package audio

import (
	"fmt"
	"sync"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	proxy "github.com/thinkonmay/thinkremote-rtchub"
	"github.com/thinkonmay/thinkremote-rtchub/listener/multiplexer"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay/opus"
	"github.com/thinkonmay/thinkremote-rtchub/util/thread"
)

type AudioPipeline struct {
	closed bool
	mut    *sync.Mutex

	clockRate float64

	codec       string
	Multiplexer *multiplexer.Multiplexer
}

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(memory *proxy.SharedMemory) (*AudioPipeline, error) {
	pipeline := &AudioPipeline{
		closed:    false,
		clockRate: 48000,
		codec:     webrtc.MimeTypeOpus,
		mut:       &sync.Mutex{},

		Multiplexer: multiplexer.NewMultiplexer("audio", opus.NewOpusPayloader()),
	}

	go func() {
		thread.HighPriorityThread()
		buffer := make([]byte, 256*1024) //256kB

		for {

			pipeline.Multiplexer.Send(buffer, uint32(pipeline.clockRate/100))
		}
	}()
	return pipeline, nil
}

func (p *AudioPipeline) GetCodec() string {
	return p.codec
}

func (p *AudioPipeline) Close() {
	fmt.Println("stoping audio pipeline")
}

func (p *AudioPipeline) RegisterRTPHandler(id string, fun func(pkt *rtp.Packet)) {
	p.Multiplexer.RegisterRTPHandler(id, fun)
}

func (p *AudioPipeline) DeregisterRTPHandler(id string) {
	p.Multiplexer.DeregisterRTPHandler(id)
}
