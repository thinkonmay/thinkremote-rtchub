package rtppay

import "github.com/pion/rtp"

type Packetizer interface {
	Packetize(payload []byte, samples uint32) []*rtp.Packet
}