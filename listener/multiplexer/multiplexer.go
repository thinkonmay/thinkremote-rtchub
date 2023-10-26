package multiplexer

import (
	"fmt"
	"sync"
	"time"

	"github.com/pion/rtp"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay"
	"github.com/thinkonmay/thinkremote-rtchub/util/win32"
)

import "C"

const (
	no_limit          = 200
	thread_count      = 1
)

type sample struct {
	data []byte 
	samples uint32
	id int
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
	buffer chan *rtp.Packet
	closed bool
}




func NewMultiplexer(id string,packetizer func() rtppay.Packetizer) *Multiplexer {
	ret := &Multiplexer{
		id:      id,
		mutex:   &sync.Mutex{},
		queue:   make(chan *sample, no_limit),
		handler: map[string]*Handler{},
		packetizer: packetizer(),
	}


	packetize := func() {
		win32.HighPriorityThread()
		for {
			sample := <-ret.queue
			packets := ret.packetizer.Packetize(sample.data,sample.samples)
			ret.mutex.Lock()
			for _,handler := range ret.handler { 
				for _, p := range packets {
					handler.buffer <- p
				}
			}
			ret.mutex.Unlock()
		}
	}

	go packetize()
	return ret
}

func (ret *Multiplexer) Send(Buff []byte, Samples uint32) {
	sample := &sample{
		data: make([]byte, len(Buff)),
		samples: Samples,
	}
	copy(sample.data,Buff)
	ret.queue <- sample
}

func (p *Multiplexer) Close() {
	keys := make([]string, 0, len(p.handler))
	for k := range p.handler {
		keys = append(keys, k)
	}
	
	for _,v := range keys {
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
		buffer: make(chan *rtp.Packet,1000),
		closed: false,
	}

	p.handler[id] = &handler
	go func() {
		win32.HighPriorityThread()
		for {
			if handler.closed {
				return
			} else if len(handler.buffer) == 0{
				time.Sleep(time.Millisecond)
				continue
			}

			handler.handler(<-handler.buffer)
		}
	}()
}

func (p *Multiplexer) DeregisterRTPHandler(id string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.handler[id].closed = true
	delete(p.handler,id)
	fmt.Printf("deregister RTP handler %s\n",id)
}