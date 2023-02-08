package listener

import (
	"github.com/pion/rtp"
	"github.com/thinkonmay/thinkremote-rtchub/util/config"
)

type Listener interface {
	GetConfig() *config.ListenerConfig

	SetSource(source interface{}) error
	SetProperty(name string,val int) error
	GetSourceName() string

	ReadRTP() *rtp.Packet

	Open()
	Close()
}
