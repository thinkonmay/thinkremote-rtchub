package udp

import (
	"fmt"
	"net"

	"github.com/OnePlay-Internet/webrtc-proxy/broadcaster"
	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	"github.com/pion/rtp"
)

type UDPBroadcaster struct {
	config *config.BroadcasterConfig
	conn   net.Conn
	port   int

	buffer     []byte
	bufferSize int

	closeChannel  chan bool
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
		packet := <-*loop.chann
		size, err := packet.MarshalTo(loop.buf)
		if err != nil {
			fmt.Printf("%v", err)
		}
		_, err = loop.conn.Write(loop.buf[:size])
		if err != nil {
			fmt.Printf("%s\n", err.Error())
		}
	}
	go loop.lop()
}

func NewUDPBroadcaster(config *config.BroadcasterConfig) (udp *UDPBroadcaster, err error) {
	udp = &UDPBroadcaster{}
	udp.config = config
	udp.bufferSize = 10000
	udp.port = config.Port
	if err != nil {
		return
	}
	udp.buffer = make([]byte, udp.bufferSize)
	udp.closeChannel = make(chan bool)
	udp.packetChannel = make(chan *rtp.Packet)
	udp.conn, err = net.Dial("udp", fmt.Sprintf("localhost:%d", config.Port))

	newloop(udp.conn, udp.packetChannel).runloop()
	newloop(udp.conn, udp.packetChannel).runloop()
	newloop(udp.conn, udp.packetChannel).runloop()
	return
}

func (udp *UDPBroadcaster) Write(packet *rtp.Packet) {
	udp.packetChannel <- packet
}

func (udp *UDPBroadcaster) Close() {
	udp.closeChannel <- true
}

func (udp *UDPBroadcaster) OnClose(fun broadcaster.OnCloseFunc) {
	go func() {
		<-udp.closeChannel
		fun(udp)
	}()
}

func (udp *UDPBroadcaster) ReadConfig() *config.BroadcasterConfig {
	return udp.config
}
