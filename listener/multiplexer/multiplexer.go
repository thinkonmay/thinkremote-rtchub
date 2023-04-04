package multiplexer

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/pion/rtp"
	"github.com/thinkonmay/thinkremote-rtchub/listener/rtppay"
)

import "C"

const (
	enable_soft_limit = false 
	soft_limit        = 40
	hard_limit        = 50
	no_limit          = 50
)


type Multiplexer struct {
	raw    chan *struct{
		buff[]byte
		samples int
	}

	packetizer rtppay.Packetizer

	mutex   *sync.Mutex
	handler map[string]Handler
}

type Handler struct {
	handler func(*rtp.Packet)
}




func NewMultiplexer(packetizer func() rtppay.Packetizer) *Multiplexer {
	ret := &Multiplexer{
		raw:     make(chan *struct{buff []byte; samples int},hard_limit),
		mutex:   &sync.Mutex{},
		handler: map[string]Handler{},
		packetizer: packetizer(),
	}



	multiply := func() {
		for {
			src_pkt := <- ret.raw
			packets := ret.packetizer.Packetize(src_pkt.buff,uint32(src_pkt.samples))
			go func() {
				for _,packet := range packets {
					for _,handler := range ret.handler {
						handler.handler(packet); 
					}
				}
			}()
		}
	}
	go multiply()
	return ret
}

func (ret *Multiplexer) Send(Buff unsafe.Pointer, bufferLen uint32,Samples uint32) {
	ret.raw<-&struct{buff []byte; samples int}{
		buff: C.GoBytes(Buff,C.int(bufferLen)),
		samples: int(Samples),
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