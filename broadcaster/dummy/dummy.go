package dummy

import (
	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	"github.com/pion/rtp"
)

type DummyBroadcaster struct {
	config *config.BroadcasterConfig
}

func NewDummyBroadcaster(config *config.BroadcasterConfig) (udp *DummyBroadcaster, err error) {
	udp = &DummyBroadcaster{
		config: config,
	}
	err = nil
	return
}

func (udp *DummyBroadcaster) Write(packet *rtp.Packet) {
}

func (udp *DummyBroadcaster) Close() {
}

func (udp *DummyBroadcaster) Open() *config.BroadcasterConfig {
	return udp.config
}
