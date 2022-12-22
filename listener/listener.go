package listener

import (
	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	"github.com/pion/rtp"
)

type Listener interface {
	GetConfig() *config.ListenerConfig

	SetSource(source interface{}) error
	SetProperty(name string,val interface{}) error
	GetSourceName() string

	ReadRTP() *rtp.Packet

	Open()
	Close()
}
