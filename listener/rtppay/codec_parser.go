package rtppay

import (
	"unsafe"

	"github.com/pion/rtp"
)

type Packetizer interface {
	Packetize(buff unsafe.Pointer, bufferLen uint32, samples uint32) []*rtp.Packet
}