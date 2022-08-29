package broadcaster

import (
	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	"github.com/pion/rtp"
)


type Broadcaster interface {
	ReadConfig() *config.BroadcasterConfig
	Write(pk *rtp.Packet)
	Close()
}
