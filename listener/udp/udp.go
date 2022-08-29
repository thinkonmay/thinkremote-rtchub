package udp

import (
	"net"
	"strconv"
	"strings"
	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	"github.com/OnePlay-Internet/webrtc-proxy/util/queue"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media"
)

type UDPListener struct {
	config *config.ListenerConfig
	conn   *net.UDPConn
	port   int64

	queue         *queue.RtpQueue
	packetChannel chan *rtp.Packet
}

func NewUDPListener(config *config.ListenerConfig) (udp UDPListener, err error) {
	udp.config = config
	udp.port, err = strconv.ParseInt(strings.Split(config.Source, ":")[1], 10, 8)
	if err != nil {
		return
	}
	udp.conn, err = net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP("localhost"),
		Port: int(udp.port),
	})
	if err != nil {
		return
	}

	udp.packetChannel = make(chan *rtp.Packet)
	udp.queue = &queue.RtpQueue{
		Outqueue:  udp.packetChannel,
		Source:    udp.conn,
		Threadnum: 15, // TODO (evaluation point)
		Bufsize:   10000,
	}

	// Read RTP packets forever and send them to the WebRTC Client
	udp.queue.Start()
	return
}


func (udp *UDPListener) ReadRTP() *rtp.Packet {
	return <-udp.packetChannel
}
func (udp *UDPListener) ReadSample() *media.Sample {
	block := make(chan *media.Sample)
	return <-block
}

func (udp *UDPListener) Close() {
	udp.queue.Closed = true
}


func (udp *UDPListener) ReadConfig() *config.ListenerConfig {
	return udp.config
}
