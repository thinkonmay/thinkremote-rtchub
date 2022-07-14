package listener

import (
	"github.com/pigeatgarlic/webrtc-proxy/util/config"
	"github.com/pion/rtp"
)

type OnCloseFunc func(lis Listener)

type Listener interface {
	ReadConfig() *config.ListenerConfig
	Open()
	Read() (*rtp.Packet)
	OnClose(fun OnCloseFunc)
	Close()
}