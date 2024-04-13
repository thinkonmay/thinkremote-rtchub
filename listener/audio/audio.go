package audio

import (
	"fmt"
	"sync"
	"time"

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
func CreatePipeline(memory *proxy.Queue) (*AudioPipeline, error) {
	pipeline := &AudioPipeline{
		closed:    false,
		clockRate: 48000,
		codec:     webrtc.MimeTypeOpus,
		mut:       &sync.Mutex{},

		Multiplexer: multiplexer.NewMultiplexer("audio", opus.NewOpusPayloader()),
	}

	go func(queue *proxy.Queue) {
		thread.HighPriorityThread()
		buffer := make([]byte, 256*1024) //256kB
		local_index := queue.CurrentIndex()

		for {
			for local_index >= queue.CurrentIndex() {
				time.Sleep(time.Millisecond)
			}

			local_index++
			size := queue.Copy(buffer, local_index)
			pipeline.Multiplexer.Send(buffer[:size], uint32(pipeline.clockRate/100))
		}
	}(memory)
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
