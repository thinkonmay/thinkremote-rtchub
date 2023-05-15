package av1

import (
	"errors"
	"time"

	"github.com/pion/randutil"
	"github.com/pion/rtp"
	"github.com/pion/rtp/pkg/obu"
)


func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

const (
	zMask     = byte(0b10000000)
	zBitshift = 7

	yMask     = byte(0b01000000)
	yBitshift = 6

	wMask     = byte(0b00110000)
	wBitshift = 4

	nMask     = byte(0b00001000)
	nBitshift = 3

	av1PayloaderHeadersize = 1
	rtpOutboundMTU = 1200
)

var (
	errShortPacket          = errors.New("packet is not large enough")
	errNilPacket            = errors.New("invalid nil packet")

	// AV1 Errors
	errIsKeyframeAndFragment = errors.New("bits Z and N are set. Not possible to have OBU be tail fragment and be keyframe")
)

// AV1Payloader payloads AV1 packets
type AV1Payloader struct{
	MTU              uint16
	PayloadType      uint8
	SSRC             uint32
	Sequencer        rtp.Sequencer
	Timestamp        uint32
	ClockRate        uint32
	extensionNumbers struct { // put extension numbers in here. If they're 0, the extension is disabled (0 is not a legal extension number)
		AbsSendTime int // http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time
	}
	timegen func() time.Time
}


// NewPacketizer returns a new instance of a Packetizer for a specific payloader
func NewAV1Payloader(mtu uint16, pt uint8, ssrc uint32, clockRate uint32) *AV1Payloader {
	return &AV1Payloader{
		MTU:         rtpOutboundMTU,
		PayloadType:      0,
		SSRC:             0,
		Sequencer:   rtp.NewRandomSequencer(),
		Timestamp:   randutil.NewMathRandomGenerator().Uint32(),
		ClockRate:   clockRate,
		extensionNumbers: struct{ AbsSendTime int }{AbsSendTime: 22},
		timegen:     time.Now,
	}
}

// Payload fragments a AV1 packet across one or more byte arrays
// See AV1Packet for description of AV1 Payload Header
func (p *AV1Payloader) payload(mtu uint16, payload []byte) (payloads [][]byte) {
	maxFragmentSize := int(mtu) - av1PayloaderHeadersize - 2
	payloadDataRemaining := len(payload)
	payloadDataIndex := 0

	// Make sure the fragment/payload size is correct
	if min(maxFragmentSize, payloadDataRemaining) <= 0 {
		return payloads
	}

	for payloadDataRemaining > 0 {
		currentFragmentSize := min(maxFragmentSize, payloadDataRemaining)
		leb128Size := 1
		if currentFragmentSize >= 127 {
			leb128Size = 2
		}

		out := make([]byte, av1PayloaderHeadersize+leb128Size+currentFragmentSize)
		leb128Value := obu.EncodeLEB128(uint(currentFragmentSize))
		if leb128Size == 1 {
			out[1] = byte(leb128Value)
		} else {
			out[1] = byte(leb128Value >> 8)
			out[2] = byte(leb128Value)
		}

		copy(out[av1PayloaderHeadersize+leb128Size:], payload[payloadDataIndex:payloadDataIndex+currentFragmentSize])
		payloads = append(payloads, out)

		payloadDataRemaining -= currentFragmentSize
		payloadDataIndex += currentFragmentSize

		if len(payloads) > 1 {
			out[0] ^= zMask
		}
		if payloadDataRemaining != 0 {
			out[0] ^= yMask
		}
	}

	return payloads
}

// AV1Packet represents a depacketized AV1 RTP Packet
//
//  0 1 2 3 4 5 6 7
// +-+-+-+-+-+-+-+-+
// |Z|Y| W |N|-|-|-|
// +-+-+-+-+-+-+-+-+
//
// https://aomediacodec.github.io/av1-rtp-spec/#44-av1-aggregation-header
type AV1Packet struct {
	// Z: MUST be set to 1 if the first OBU element is an
	//    OBU fragment that is a continuation of an OBU fragment
	//    from the previous packet, and MUST be set to 0 otherwise.
	Z bool

	// Y: MUST be set to 1 if the last OBU element is an OBU fragment
	//    that will continue in the next packet, and MUST be set to 0 otherwise.
	Y bool

	// W: two bit field that describes the number of OBU elements in the packet.
	//    This field MUST be set equal to 0 or equal to the number of OBU elements
	//    contained in the packet. If set to 0, each OBU element MUST be preceded by
	//    a length field. If not set to 0 (i.e., W = 1, 2 or 3) the last OBU element
	//    MUST NOT be preceded by a length field. Instead, the length of the last OBU
	//    element contained in the packet can be calculated as follows:
	// Length of the last OBU element =
	//    length of the RTP payload
	//  - length of aggregation header
	//  - length of previous OBU elements including length fields
	W byte

	// N: MUST be set to 1 if the packet is the first packet of a coded video sequence, and MUST be set to 0 otherwise.
	N bool

	// Each AV1 RTP Packet is a collection of OBU Elements. Each OBU Element may be a full OBU, or just a fragment of one.
	// AV1Frame provides the tools to construct a collection of OBUs from a collection of OBU Elements
	OBUElements [][]byte
}

// Unmarshal parses the passed byte slice and stores the result in the AV1Packet this method is called upon
func (p *AV1Packet) Unmarshal(payload []byte) ([]byte, error) {
	if payload == nil {
		return nil, errNilPacket
	} else if len(payload) < 2 {
		return nil, errShortPacket
	}

	p.Z = ((payload[0] & zMask) >> zBitshift) != 0
	p.Y = ((payload[0] & yMask) >> yBitshift) != 0
	p.N = ((payload[0] & nMask) >> nBitshift) != 0
	p.W = (payload[0] & wMask) >> wBitshift

	if p.Z && p.N {
		return nil, errIsKeyframeAndFragment
	}

	currentIndex := uint(1)
	p.OBUElements = [][]byte{}

	var (
		obuElementLength, bytesRead uint
		err                         error
	)
	for i := 1; ; i++ {
		if currentIndex == uint(len(payload)) {
			break
		}

		// If W bit is set the last OBU Element will have no length header
		if byte(i) == p.W {
			bytesRead = 0
			obuElementLength = uint(len(payload)) - currentIndex
		} else {
			obuElementLength, bytesRead, err = obu.ReadLeb128(payload[currentIndex:])
			if err != nil {
				return nil, err
			}
		}

		currentIndex += bytesRead
		if uint(len(payload)) < currentIndex+obuElementLength {
			return nil, errShortPacket
		}
		p.OBUElements = append(p.OBUElements, payload[currentIndex:currentIndex+obuElementLength])
		currentIndex += obuElementLength
	}

	return payload[1:], nil
}




// Packetize packetizes the payload of an RTP packet and returns one or more RTP packets
func (p *AV1Payloader) Packetize(payload []byte, samples uint32) []*rtp.Packet {
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