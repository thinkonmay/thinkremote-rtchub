package broadcaster

import (
	"github.com/pigeatgarlic/webrtc-proxy/util/config"
	"github.com/pion/rtp"
)

type OnCloseFunc func(lis Broadcaster)

type Broadcaster interface {
	ReadConfig() *config.BroadcasterConfig
	Write(pk *rtp.Packet)
	OnClose(fun OnCloseFunc)
	Close()
}