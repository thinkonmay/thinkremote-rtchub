package listener

import (
	"github.com/pion/rtp"
)

type Listener interface {
	GetCodec() string
	SetProperty(name string,val int) error

	ReadRTP() *rtp.Packet

	Open()
	Close()
}
