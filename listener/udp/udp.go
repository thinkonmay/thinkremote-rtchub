package udp

import (
	"net"
	"time"

	"github.com/pigeatgarlic/webrtc-proxy/listener"
	"github.com/pigeatgarlic/webrtc-proxy/util/config"
	"github.com/pigeatgarlic/webrtc-proxy/util/queue"
	"github.com/pion/rtp"
)

type UDPListener struct {
	config *config.ListenerConfig
	conn *net.UDPConn
	port int

	queue *queue.RtpQueue
	packetChannel chan *rtp.Packet
}


func NewUDPListener(config *config.ListenerConfig) (udp UDPListener, err error) {
	udp.config = config;
	udp.port = config.Port;
	udp.conn,err = net.ListenUDP("udp", &net.UDPAddr {
		IP: net.ParseIP("localhost"), 
		Port: udp.port, 
	});
	if err != nil {
		return;
	}

	udp.packetChannel = make(chan *rtp.Packet);
	udp.queue = &queue.RtpQueue{
		Outqueue: udp.packetChannel,
		Source: udp.conn,
		Threadnum: 15, // TODO (evaluation point)
		Bufsize: udp.config.BufferSize,
	}

	return;
}

func (udp *UDPListener)	Open() {
	// Read RTP packets forever and send them to the WebRTC Client
	udp.queue.Start();
}

func (udp *UDPListener) Read() *rtp.Packet {
	return <-udp.packetChannel;
}

func (udp *UDPListener)	Close() {
	udp.queue.Closed = true;
}

func (udp *UDPListener)	OnClose(fun listener.OnCloseFunc) {
	go func() {
		if udp.queue.Closed {
			fun(udp);	
		}
		time.Sleep(100 * time.Millisecond)
	}()
}
 
func (udp *UDPListener)	ReadConfig() *config.ListenerConfig{
	return udp.config;
}