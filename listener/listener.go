package listener

import (
	"github.com/pion/rtp"
	"github.com/thinkonmay/thinkremote-rtchub/util/config"
)

type Listener interface {
	GetConfig() *config.ListenerConfig
	SetProperty(name string,val int) error

	ReadRTP() *rtp.Packet

	Open()
	Close()
}
