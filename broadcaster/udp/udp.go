package udp

import (
	"fmt"
	"net"

	"github.com/pigeatgarlic/webrtc-proxy/broadcaster"
	"github.com/pigeatgarlic/webrtc-proxy/util/config"
)

type UDPBroadcaster struct {
	config *config.BroadcasterConfig
	conn net.Conn
	port int

	buffer []byte
	bufferSize int

	closeChannel chan bool
}


func NewUDPBroadcaster(config *config.BroadcasterConfig) (udp *UDPBroadcaster, err error) {
	udp = &UDPBroadcaster{}
	udp.config = config;
	udp.bufferSize = config.BufferSize;
	udp.port = config.Port;
	if err != nil {
		return;
	}
	udp.buffer = make([]byte, udp.bufferSize);
	udp.closeChannel = make(chan bool);
	udp.conn,err = net.Dial("udp",fmt.Sprintf("localhost:%d",config.Port));
	return;
}



func (udp *UDPBroadcaster) Write(size int, data []byte) error {
	fmt.Printf("writting buffer size %d\n",size);
	str := string(data[:size]);
	written,err := fmt.Fprintf(udp.conn,str);
	if written != size {
		err = fmt.Errorf("wrong size");
	}
	return err;
}

func (udp *UDPBroadcaster)	Close() {
	udp.closeChannel <- true;
}

func (udp *UDPBroadcaster)	OnClose(fun broadcaster.OnCloseFunc) {
	go func() {
		<-udp.closeChannel;
		fun(udp);	
	}()
}
 


func (udp *UDPBroadcaster)	ReadConfig() *config.BroadcasterConfig{
	return udp.config;
}