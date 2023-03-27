package multiplexer

import (
	"fmt"
	"sync"

	"github.com/pion/rtp"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay"
)

const (
	enable_soft_limit = true
	soft_limit        = 40
	hard_limit        = 50
	no_limit          = 500
)

type RawData struct {
	Samples uint32
	Buff []byte
}
type Multiplexer struct {
	srcPkt chan *rtp.Packet
	In     chan *RawData

	packetizer rtppay.Packetizer

	mutex   *sync.Mutex
	handler map[string]Handler
}

type Handler struct {
	closed  bool
	sink    chan *rtp.Packet
	handler func(*rtp.Packet)
}




func NewMultiplexer(packetizer func() rtppay.Packetizer) *Multiplexer {
	ret := &Multiplexer{
		srcPkt:  make(chan *rtp.Packet, hard_limit),
		In:      make(chan *RawData,hard_limit),
		mutex:   &sync.Mutex{},
		handler: map[string]Handler{},
		packetizer: packetizer(),
	}



	go func() {
		for {
			src_buffer := <- ret.In
			packets := ret.packetizer.Packetize(src_buffer.Buff,src_buffer.Samples)
			for _, packet := range packets {
				ret.srcPkt <- packet
			}
		}
	}()

	go func() {
		for {
			src_pkt := <- ret.srcPkt
			for _,v := range ret.handler {
				if len(ret.srcPkt) > soft_limit && enable_soft_limit {
					continue
				}
				v.sink <- src_pkt
			}
		}
	}()
	return ret
}

func (p *Multiplexer) Send( Buff []byte,Samples uint32) {
	p.In<-&RawData{
		Samples: Samples,
		Buff: Buff,
	}
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

	sink := make(chan *rtp.Packet,no_limit)
	p.handler[id] = Handler{
		closed: false,
		sink: sink,
		handler: fun,
	}

	go func ()  { for { fun(<-sink); } }()
}

func (p *Multiplexer) DeregisterRTPHandler(id string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	handler := p.handler[id];
	handler.closed = true

	delete(p.handler,id)
	fmt.Printf("deregister RTP handler %s\n",id)
}