package multiplexer

import (
	"fmt"
	"sync"

	"github.com/pion/rtp"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay"
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
	handler map[string]Handler
}

type Handler struct {
	handler func(*rtp.Packet)
}




func NewMultiplexer(id string,packetizer func() rtppay.Packetizer) *Multiplexer {
	ret := &Multiplexer{
		id:      id,
		mutex:   &sync.Mutex{},
		queue:   make(chan *sample, no_limit),
		handler: map[string]Handler{},
		packetizer: packetizer(),
	}


	sender := func (handler Handler,packets []*rtp.Packet) {
		for _,packet := range packets {
			handler.handler(packet); 
		}
	}

	packetize := func() {
		for {
			sample := <-ret.queue
			packets := ret.packetizer.Packetize(sample.data,sample.samples)
			for _,handler := range ret.handler {
				go sender(handler,packets)
			}
		}
	}

	for i := 0; i < thread_count; i++ {
		go packetize()
	}
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
	p.handler[id] = Handler{
		handler: fun,
	}
}

func (p *Multiplexer) DeregisterRTPHandler(id string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	delete(p.handler,id)
	fmt.Printf("deregister RTP handler %s\n",id)
}