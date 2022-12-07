package listener

import (
	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	"github.com/pion/rtp"
)

type Listener interface {
	GetConfig() *config.ListenerConfig 
	UpdateConfig(config *config.ListenerConfig) error

	ReadRTP() *rtp.Packet
	
	Open() 
	Close()
}
