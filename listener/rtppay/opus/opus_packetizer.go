package opus

import (
	"time"

	"github.com/pion/randutil"
	"github.com/pion/rtp"
)

const (
	// Unknown defines default public constant to use for "enum" like struct
	// comparisons when no value was defined.
	Unknown    = iota
	unknownStr = "unknown"

	rtpOutboundMTU = 1200
)

// H264Payloader payloads H264 packets
type OPUSPayloader struct {
	MTU              uint16
	PayloadType      uint8
	SSRC             uint32
	Sequencer        rtp.Sequencer
	Timestamp        uint32
	extensionNumbers struct { // put extension numbers in here. If they're 0, the extension is disabled (0 is not a legal extension number)
		AbsSendTime int // http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time
	}
	timegen func() time.Time
}

func NewOpusPayloader() *OPUSPayloader {
	return &OPUSPayloader{
		MTU: rtpOutboundMTU,
		PayloadType: 0,
		SSRC: 0,
		Sequencer: rtp.NewRandomSequencer(),
		Timestamp: randutil.NewMathRandomGenerator().Uint32(),
		extensionNumbers: struct{AbsSendTime int}{AbsSendTime: 22},
		timegen: time.Now,
	}
}


// payload fragments an Opus packet across one or more byte arrays
func (p *OPUSPayloader) payload(mtu uint16, payload []byte) [][]byte {
	if payload == nil {
		return [][]byte{}
	}

	out := make([]byte, len(payload))
	copy(out, payload)
	return [][]byte{out}
}

// Packetize packetizes the payload of an RTP packet and returns one or more RTP packets
func (p *OPUSPayloader) Packetize(payload []byte, samples uint32) []*rtp.Packet {
	payloads := p.payload(p.MTU-12, payload)
	packets := make([]*rtp.Packet, len(payloads))

	for i, pp := range payloads {
		packets[i] = &rtp.Packet{
			Header: rtp.Header{
				Version:        2,
				Padding:        false,
				Extension:      false,
				Marker:         i == len(payloads)-1,
				PayloadType:    p.PayloadType,
				SequenceNumber: p.Sequencer.NextSequenceNumber(),
				Timestamp:      p.Timestamp, // Figure out how to do timestamps
				SSRC:           p.SSRC,
			},
			Payload: pp,
		}
	}
	p.Timestamp += samples

	if len(packets) != 0 && p.extensionNumbers.AbsSendTime != 0 {
		sendTime := rtp.NewAbsSendTimeExtension(p.timegen())
		// apply http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time
		b, err := sendTime.Marshal()
		if err != nil {
			return nil // never happens
		}
		err = packets[len(packets)-1].SetExtension(uint8(p.extensionNumbers.AbsSendTime), b)
		if err != nil {
			return nil // never happens
		}
	}

	return packets
}