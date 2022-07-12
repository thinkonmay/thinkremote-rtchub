package broadcaster

import "github.com/pigeatgarlic/webrtc-proxy/util/config"

type OnCloseFunc func(lis Broadcaster)

type Broadcaster interface {
	ReadConfig() *config.BroadcasterConfig
	Write(size int, data []byte) error
	OnClose(fun OnCloseFunc)
	Close()
}