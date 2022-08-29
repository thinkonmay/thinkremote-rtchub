package listener

import (
	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media"
)

type Listener interface {
	ReadConfig() *config.ListenerConfig
	ReadRTP() *rtp.Packet
	ReadSample() *media.Sample
	Close()
}
