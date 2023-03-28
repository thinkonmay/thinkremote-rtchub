package rtppay

import (
	"github.com/pion/rtp"
)

type Packetizer interface {
	Packetize(buff []byte, samples uint32) []*rtp.Packet
}