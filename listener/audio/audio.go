package audio

import (
	"sync"
	"time"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
	proxy "github.com/thinkonmay/thinkremote-rtchub"
	"github.com/thinkonmay/thinkremote-rtchub/listener/multiplexer"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay/opus"
	"github.com/thinkonmay/thinkremote-rtchub/util/thread"
)

type AudioPipeline struct {
	closed chan bool
	mut    *sync.Mutex

	clockRate float64

	codec       string
	Multiplexer *multiplexer.Multiplexer
}

func CreatePipeline(queue *proxy.Queue) (*AudioPipeline, error) {
	pipeline := &AudioPipeline{
		closed:    make(chan bool, 2),
		clockRate: 48000,
		codec:     webrtc.MimeTypeOpus,
		mut:       &sync.Mutex{},

		Multiplexer: multiplexer.NewMultiplexer("audio", opus.NewOpusPayloader()),
	}

	buffer := make([]byte, 256*1024) //256kB
	local_index := queue.CurrentIndex()
	thread.HighPriorityLoop(pipeline.closed, func() {
		for local_index >= queue.CurrentIndex() {
			time.Sleep(time.Millisecond)
		}

		local_index++
		size, _ := queue.Copy(buffer, local_index)
		pipeline.Multiplexer.Send(buffer[:size], uint32(pipeline.clockRate/100))
	})
	return pipeline, nil
}

func (p *AudioPipeline) GetCodec() string {
	return p.codec
}

func (p *AudioPipeline) Close() {
	thread.TriggerStop(p.closed)
}

func (p *AudioPipeline) RegisterRTPHandler(id string, fun func(pkt *rtp.Packet)) {
	p.Multiplexer.RegisterRTPHandler(id, fun)
}

func (p *AudioPipeline) DeregisterRTPHandler(id string) {
	p.Multiplexer.DeregisterRTPHandler(id)
}
