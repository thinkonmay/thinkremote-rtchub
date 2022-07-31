package queue

import (
	"fmt"
	"github.com/pion/rtp"
)

type rtpsource interface {
	Read([]byte) (size int, err error)
}

type buffer struct {
	buf []byte
	size int
	inuse int
}

type RtpQueue struct {
	inchan chan *buffer
	outchan chan *buffer

	Closed bool
	Outqueue chan *rtp.Packet
	Source rtpsource
	Threadnum int
	Bufsize int
}


func (queue *RtpQueue) Start(){
	queue.inchan  = make(chan *buffer)
	queue.outchan = make(chan *buffer)
	queue.Closed = false;

	go func() {
		for i := 0; i < queue.Threadnum; i++ {
			queue.inchan<- &buffer{
				buf: make([]byte,queue.Bufsize),
				size: queue.Bufsize,
				inuse: 0,
			}
		}
	}()


	go func() {
		for {
			queue.load()	
		}
	}()

	for i := 0; i < queue.Threadnum; i++ {
		go func() {
			for {
				queue.transform()
			}	
		}()
	}
}



func (queue *RtpQueue) load() {
	var err error;
	buffer := <-queue.inchan
	buffer.inuse, err = queue.Source.Read(buffer.buf)
	if err != nil{
		fmt.Printf("fail to read from source: %s\n",err.Error());	
		return;
	}
	queue.outchan <-buffer;
}

func (queue *RtpQueue) transform() {
	buffer := <-queue.outchan
	packet := rtp.Packet{}
	err := packet.Unmarshal(buffer.buf[:buffer.inuse]);
	if err != nil {
		fmt.Printf("fail to unmarshal rtp packet: %s\n",err.Error());	
	}
	queue.Outqueue<-&packet;
	buffer.inuse = 0;
	queue.inchan<-buffer;
}