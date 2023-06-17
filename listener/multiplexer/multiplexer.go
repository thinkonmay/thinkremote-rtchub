package multiplexer

import (
	"fmt"
	"sync"

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
	id string



	packetizer rtppay.Packetizer

	mutex   *sync.Mutex
	handler map[string]Handler
}

type Handler struct {
	handler func(*rtp.Packet)
}




func NewMultiplexer(id string,packetizer func() rtppay.Packetizer) *Multiplexer {
	ret := &Multiplexer{
		id:      id,
		mutex:   &sync.Mutex{},
		handler: map[string]Handler{},
		packetizer: packetizer(),
	}

	return ret
}

func (ret *Multiplexer) Send(Buff []byte, Samples uint32) {
	go func() {
		packets := ret.packetizer.Packetize(Buff,uint32(Samples))
		for _,handler := range ret.handler {
			for _,packet := range packets {
				handler.handler(packet); 
			}
		}
	}()
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