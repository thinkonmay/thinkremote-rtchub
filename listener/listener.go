package listener

import (
	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media"
)

type Listener interface {
	GetConfig() *config.ListenerConfig 
	UpdateConfig( config *config.ListenerConfig ) error

	ReadRTP() *rtp.Packet
	ReadSample() *media.Sample

	Open() *config.ListenerConfig
	Close()
}
