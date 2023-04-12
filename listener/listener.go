package listener

import (
	"github.com/pion/rtp"
)

type Listener interface {
	GetCodec() string
	SetProperty(name string,val int) error
	RegisterRTPHandler(string,func(*rtp.Packet)) 
	DeregisterRTPHandler(string) 

	Open() 
	Close()
}
