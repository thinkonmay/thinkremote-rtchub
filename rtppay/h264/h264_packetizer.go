package h264

import (
	"encoding/binary"
	"time"

	"github.com/OnePlay-Internet/webrtc-proxy/util/io"
	"github.com/pion/randutil"
	"github.com/pion/rtp"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

const (
	// Unknown defines default public constant to use for "enum" like struct
	// comparisons when no value was defined.
	Unknown    = iota
	unknownStr = "unknown"

	rtpOutboundMTU = 1200


	
)

// https://stackoverflow.com/questions/24884827/possible-locations-for-sequence-picture-parameter-sets-for-h-264-stream/24890903#24890903
const (
	stapaNALUType  = 24		  // 00011000
	fubNALUType    = 29		  // 00011101
	spsNALUType    = 7		  // 00000111
	ppsNALUType    = 8		  // 00001000
	audNALUType    = 9		  // 00001001
	fillerNALUType = 12		  // 00001100
	fuaNALUType    = 28		  // 00011100

	fuaHeaderSize       = 2
	stapaHeaderSize     = 1
	stapaNALULengthSize = 2

	naluTypeBitmask   = 0x1F  // 00011111
	naluRefIdcBitmask = 0x60  // 01100000
	fuStartBitmask    = 0x80  // 10000000
	fuEndBitmask      = 0x40  // 01000000
	outputStapAHeader = 0x78  // 01111000
)

const (
	NALU_NONE = 0
	NALU_RAW = 1
	NALU_AVCC = 2
	NALU_ANNEXB = 3
)


// H264Payloader payloads H264 packets
type H264Payloader struct {
	spsNalu, ppsNalu []byte


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

func NewH264Payloader() *H264Payloader {
	return &H264Payloader{
		MTU: rtpOutboundMTU,
		PayloadType: 0,
		SSRC: 0,
		Sequencer: rtp.NewRandomSequencer(),
		Timestamp: randutil.NewMathRandomGenerator().Uint32(),
		extensionNumbers: struct{AbsSendTime int}{AbsSendTime: 22},
		timegen: time.Now,
	}
}



func findIndicator(nalu []byte, start int) (indStart int, indLen int) {
	zCount := 0 // zeroCount

	for i, b := range nalu[start:] {
		if b == 0 {
			zCount++
			continue
		} else if b == 1 {
			if zCount >= 2 {
				return start + i - zCount, zCount + 1
			}
		}
		zCount = 0
	}
	return -1, -1 // return -1 if no indicator found
}




func splitNALUs(payload []byte) (nalus [][]byte, typ int) {
	typ = NALU_NONE
	nalus = [][]byte{}
	defer func ()  {
		if len(nalus) == 0 || typ == NALU_NONE {
			nalus = append(nalus, payload)
			typ = NALU_RAW
		}
	}()

	if len(payload) < 4 {
		return
	}

	// is Annex B
	// +----------------------------------------+
	// |0|1|2|3|4|5|6|7|8|9|10|11|12|13|14|15|16|...
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-
	// |x|x|0|0|0|0|1|x|x|x|x|x|x|x|x|x|x|x|x|x|...
	// |<--Start   					   		   |
	// |   |indLen   |     			nalu	   |
	// |   |<--indStart   					   |
	// +---------------------------------------+
	nextIndStart, nextIndLen := findIndicator(payload, 0)
	if nextIndStart != -1 {
		for nextIndStart != -1 {
			prevStart := nextIndStart + nextIndLen
			nextIndStart, nextIndLen = findIndicator(payload, prevStart)
			if nextIndStart != -1 {
				// emit on nalu
				nalus = append(nalus, payload[prevStart:nextIndStart])
			} else {
				// Emit until end of stream, no end indicator found
				nalus = append(nalus, payload[prevStart:])
			}
		}

		typ = NALU_ANNEXB;
		return
	} 


	// is AVCC
	// +----------------------------------------+
	// |0|1|2|3|4|5|6|7|8|9|10|11|12|13|14|15|16|...
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-
	// |x|x|s|s|s|s|x|x|x|x|x|x|x|x|x|x|x|x|x|x|...
	// |   |naluSz |     			nalu	   |
	// +---------------------------------------+
	val4 := io.U32BE(payload)
	if val4 <= uint32(len(payload)) {
		naluSz := val4
		nalu := payload[4:]

		for {
			if naluSz > uint32(len(nalu)) { // break if nalu size is greater than size of payload left
				break
			}

			nalus = append(nalus, nalu[:naluSz])  // append slice from payload with naluSz size
			nalu = nalu[naluSz:]  // move pointer naluSz byte to the left
			if len(nalu) < 4 { // break if size of nalu left smaller than 4 (sizeof uint32)
				break
			}

			naluSz = io.U32BE(nalu) // get naluSize from 4 first bytes left
			nalu = nalu[4:] // shift pointer 4 byte to the left
		}

		if len(nalu) == 0 {
			return nalus, NALU_AVCC
		} else {
			// reset nalus
			nalus = [][]byte{};
		}
	}
	return;
}



// payload fragments a H264 packet across one or more byte arrays
func (p *H264Payloader) payload(mtu uint16, payload []byte) [][]byte {
	var payloads [][]byte
	if len(payload) == 0 {
		return payloads
	}

	nalus,_ := splitNALUs(payload)
	for _,nalu := range nalus {
		if len(nalu) == 0 {
			continue
		}

		naluType   := nalu[0] & naluTypeBitmask     // AND operator on first byte of nal Unit
		naluRefIdc := nalu[0] & naluRefIdcBitmask   // AND operator on first byte of nal Unit

		switch {
		case naluType == audNALUType || naluType == fillerNALUType:
			continue
		case naluType == spsNALUType:
			p.spsNalu = nalu
			continue
		case naluType == ppsNALUType:
			p.ppsNalu = nalu
			continue
		case p.spsNalu != nil && p.ppsNalu != nil:
			// Pack current NALU with SPS and PPS as STAP-A
			spsLen := make([]byte, 2)
			binary.BigEndian.PutUint16(spsLen, uint16(len(p.spsNalu)))

			ppsLen := make([]byte, 2)
			binary.BigEndian.PutUint16(ppsLen, uint16(len(p.ppsNalu)))

			stapANalu := []byte{outputStapAHeader}
			stapANalu = append(stapANalu, spsLen...)
			stapANalu = append(stapANalu, p.spsNalu...)
			stapANalu = append(stapANalu, ppsLen...)
			stapANalu = append(stapANalu, p.ppsNalu...)
			if len(stapANalu) <= int(mtu) {
				out := make([]byte, len(stapANalu))
				copy(out, stapANalu)
				payloads = append(payloads, out)
			}

			p.spsNalu = nil
			p.ppsNalu = nil
		}

		// Single NALU
		if len(nalu) <= int(mtu) {
			out := make([]byte, len(nalu))
			copy(out, nalu)
			payloads = append(payloads, out)
			continue
		}

		// FU-A
		maxFragmentSize := int(mtu) - fuaHeaderSize

		// The FU payload consists of fragments of the payload of the fragmented
		// NAL unit so that if the fragmentation unit payloads of consecutive
		// FUs are sequentially concatenated, the payload of the fragmented NAL
		// unit can be reconstructed.  The NAL unit type octet of the fragmented
		// NAL unit is not included as such in the fragmentation unit payload,
		// but rather the information of the NAL unit type octet of the
		// fragmented NAL unit is conveyed in the F and NRI fields of the FU
		// indicator octet of the fragmentation unit and in the type field of
		// the FU header.  An FU payload MAY have any number of octets and MAY
		// be empty.

		naluData := nalu
		// According to the RFC, the first octet is skipped due to redundant information
		naluDataIndex := 1
		naluDataLength := len(nalu) - naluDataIndex
		naluDataRemaining := naluDataLength

		if min(maxFragmentSize, naluDataRemaining) <= 0 {
			continue
		}

		for naluDataRemaining > 0 {
			currentFragmentSize := min(maxFragmentSize, naluDataRemaining)
			out := make([]byte, fuaHeaderSize+currentFragmentSize)

			// +---------------+
			// |0|1|2|3|4|5|6|7|
			// +-+-+-+-+-+-+-+-+
			// |F|NRI|  Type   |
			// +---------------+
			out[0] = fuaNALUType
			out[0] |= naluRefIdc

			// +---------------+
			// |0|1|2|3|4|5|6|7|
			// +-+-+-+-+-+-+-+-+
			// |S|E|R|  Type   |
			// +---------------+

			out[1] = naluType
			if naluDataRemaining == naluDataLength {
				// Set start bit
				out[1] |= 1 << 7
			} else if naluDataRemaining-currentFragmentSize == 0 {
				// Set end bit
				out[1] |= 1 << 6
			}

			copy(out[fuaHeaderSize:], naluData[naluDataIndex:naluDataIndex+currentFragmentSize])
			payloads = append(payloads, out)

			naluDataRemaining -= currentFragmentSize
			naluDataIndex += currentFragmentSize
		}
	}

	return payloads
}





// Packetize packetizes the payload of an RTP packet and returns one or more RTP packets
func (p *H264Payloader) Packetize(payload []byte, samples uint32) []*rtp.Packet {
	// Guard against an empty payload
	if len(payload) == 0 {
		return nil
	}

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

