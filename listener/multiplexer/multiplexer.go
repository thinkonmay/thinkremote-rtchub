package multiplexer

import (
	"fmt"
	"sync"

	"github.com/pion/rtp"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay"
	"github.com/thinkonmay/thinkremote-rtchub/util/win32"
)

import "C"

const (
	queue_size     = 1024
)

type sample struct {
	data    []byte
	samples uint32
	id      int
}
type Multiplexer struct {
	id string

	packetizer rtppay.Packetizer

	mutex   *sync.Mutex
	queue   chan *sample
	handler map[string]*Handler
}

type Handler struct {
	handler func(*rtp.Packet)
	buffer  chan *rtp.Packet
}

func NewMultiplexer(id string, packetizer rtppay.Packetizer) *Multiplexer {
	ret := &Multiplexer{
		id:         id,
		mutex:      &sync.Mutex{},
		queue:      make(chan *sample, queue_size),
		handler:    map[string]*Handler{},
		packetizer: packetizer,
	}

	return ret
}

func (ret *Multiplexer) Send(Buff []byte, Samples uint32) {
	packets := ret.packetizer.Packetize(Buff, Samples)
	ret.mutex.Lock()
	defer ret.mutex.Unlock()
	for _, handler := range ret.handler {
		for _, p := range packets {
			handler.buffer <- p
		}
	}
}

func (p *Multiplexer) Close() {
	keys := make([]string, 0, len(p.handler))
	for k := range p.handler {
		keys = append(keys, k)
	}

	for _, v := range keys {
		p.DeregisterRTPHandler(v)
	}
}

func (p *Multiplexer) RegisterRTPHandler(id string, fun func(pkt *rtp.Packet)) {
	if p.handler == nil {
		fmt.Println("Try to register RTP handler while pipeline not ready")
		return
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()
	handler := Handler{
		handler: fun,
		buffer:  make(chan *rtp.Packet, queue_size),
	}

	p.handler[id] = &handler
	go func() { win32.HighPriorityThread()
		for {
			buffer := <-handler.buffer
			if buffer == nil {
				return
			}

			handler.handler(buffer)
		}
	}()
}

func (p *Multiplexer) DeregisterRTPHandler(id string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.handler[id].buffer <- nil
	delete(p.handler, id)
	fmt.Printf("deregister RTP handler %s\n", id)
}
