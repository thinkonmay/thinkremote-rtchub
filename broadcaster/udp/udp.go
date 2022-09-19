package udp

import (
	"fmt"
	"net"

	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	"github.com/pion/rtp"
)

type UDPBroadcaster struct {
	config *config.BroadcasterConfig
	conn   net.Conn
	port   int

	buffer     []byte
	bufferSize int

	packetChannel chan *rtp.Packet
}

type loop func()

type writeLoop struct {
	conn  net.Conn
	chann *chan *rtp.Packet

	lop loop
	buf []byte
}

func newloop(conn net.Conn, channel chan *rtp.Packet) (writeloop *writeLoop) {
	writeloop = &writeLoop{}
	writeloop.conn = conn
	writeloop.chann = &channel
	writeloop.buf = make([]byte, 10000)
	return
}

func (loop *writeLoop) runloop() {
	loop.lop = func() {
		for {
			packet := <-*loop.chann
			fmt.Printf("%d\n", packet.Header.SequenceNumber)
		}
	}
	go loop.lop()
}

func NewUDPBroadcaster(config *config.BroadcasterConfig) (udp *UDPBroadcaster, err error) {
	udp = &UDPBroadcaster{}
	udp.config = config
	udp.bufferSize = 10000
	udp.port = 6000
	if err != nil {
		return
	}
	udp.buffer = make([]byte, udp.bufferSize)
	udp.packetChannel = make(chan *rtp.Packet)
	udp.conn, err = net.Dial("udp", fmt.Sprintf("localhost:%d", udp.port))

	newloop(udp.conn, udp.packetChannel).runloop()
	return
}

func (udp *UDPBroadcaster) Write(packet *rtp.Packet) {
	udp.packetChannel <- packet
}

func (udp *UDPBroadcaster) Close() {
}


func (udp *UDPBroadcaster) Open() *config.BroadcasterConfig {
	return udp.config
}
