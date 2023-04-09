package multiplexer

import (
	"fmt"
	"sync"
	"time"
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
	id string

	recv_sample_count int
	send_count int
	total_packetize_nanosecond int

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




func NewMultiplexer(id string,packetizer func() rtppay.Packetizer) *Multiplexer {
	ret := &Multiplexer{
		id:      id,
		recv_sample_count: 0,
		send_count: 0,
		total_packetize_nanosecond: 0,

		raw:     make(chan *struct{buff []byte; samples int},hard_limit),
		mutex:   &sync.Mutex{},
		handler: map[string]Handler{},
		packetizer: packetizer(),
	}


	go func() {
		for {
			time.Sleep(10 * time.Second)
			fmt.Printf("[%s] multiplexer %s report : %d samples received, %d packets sent, total packetize time: %dns",time.Now().Format(time.RFC3339), ret.id,ret.recv_sample_count,ret.send_count,ret.total_packetize_nanosecond)
			ret.recv_sample_count = 0
			ret.send_count = 0
			ret.total_packetize_nanosecond = 0
		}
	}()

	multiply := func() {
		for {
			src_pkt := <- ret.raw

			pre := time.Now().UnixNano()
			packets := ret.packetizer.Packetize(src_pkt.buff,uint32(src_pkt.samples))
			ret.total_packetize_nanosecond = ret.total_packetize_nanosecond + int(time.Now().UnixNano() - pre)

			go func() {
				for _,handler := range ret.handler {
					for _,packet := range packets {
						handler.handler(packet); 
						ret.send_count++
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
	ret.recv_sample_count = ret.recv_sample_count + int(Samples)
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