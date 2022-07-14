package udp

import (
	"fmt"
	"net"

	"github.com/pigeatgarlic/webrtc-proxy/listener"
	"github.com/pigeatgarlic/webrtc-proxy/util/config"
	"github.com/pion/rtp"
)

type UDPListener struct {
	config *config.ListenerConfig
	conn *net.UDPConn
	port int

	buffer []byte
	bufferSize int

	packetChannel chan *rtp.Packet
	closeChannel chan bool
	closed bool
}


func NewUDPListener(config *config.ListenerConfig) (udp UDPListener, err error) {
	udp.config = config;
	udp.bufferSize = config.BufferSize;
	udp.port = config.Port;
	udp.conn,err = net.ListenUDP("udp", &net.UDPAddr {
		IP: net.ParseIP("localhost"), 
		Port: udp.port, 
	});
	if err != nil {
		return;
	}
	udp.buffer = make([]byte, udp.bufferSize);
	udp.closeChannel = make(chan bool);
	udp.packetChannel = make(chan *rtp.Packet);
	udp.closed = true;
	return;
}

func (udp *UDPListener)	Open() {
	// Read RTP packets forever and send them to the WebRTC Client
	udp.closed = false;
	go func() {
		defer func(){
			udp.closeChannel <- true;	
		}();

		for {
			size, _, err := udp.conn.ReadFrom(udp.buffer)
			if err != nil {
				fmt.Printf("udp error: %s\n",err)
				continue;
			}
			if udp.closed {
				return;
			}
			pk := rtp.Packet{}

			pk.Unmarshal(udp.buffer[:size])

			fmt.Printf("got %dbyte from port %d\n", size, udp.port)
			fmt.Printf("GOT %s\n", pk.String())

			udp.packetChannel <- &pk;
		}
	}();
}

func (udp *UDPListener) Read() *rtp.Packet {
	return <-udp.packetChannel;
}

func (udp *UDPListener)	Close() {
	udp.closeChannel <- true;
}

func (udp *UDPListener)	OnClose(fun listener.OnCloseFunc) {
	go func() {
		<-udp.closeChannel;
		udp.closed = true;
		fun(udp);	
	}()
}
 


func (udp *UDPListener)	ReadConfig() *config.ListenerConfig{
	return udp.config;
}