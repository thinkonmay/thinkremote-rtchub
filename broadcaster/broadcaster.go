package broadcaster

import (
	"github.com/pion/rtp"
)

type Broadcaster interface {
	Write(pk *rtp.Packet)
	Close()
}
