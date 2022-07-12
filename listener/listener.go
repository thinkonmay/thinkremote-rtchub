package listener

import "github.com/pigeatgarlic/webrtc-proxy/util/config"

type OnCloseFunc func(lis Listener)

type Listener interface {
	ReadConfig() *config.ListenerConfig
	Open()
	Read() (size int, data []byte)
	OnClose(fun OnCloseFunc)
	Close()
}