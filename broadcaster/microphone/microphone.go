package microphone

import (
	"io"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/pion/opus"
)

func StartMicrophone(data chan *[]byte) {
	format := beep.Format{
		SampleRate:  beep.SampleRate(48000),
		NumChannels: 2,
		Precision:   1,
	}

	reader := opusReader{
		decodeBuffer: make([]byte, 1920),
		opusDecoder:  opus.NewDecoder(),
		input: data,
	}

	stream := pcmStream{
		reader:   &reader,
		format:   format,
		buf: make([]byte, 512*format.Width()),
	}
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	speaker.Play(&stream)
}

func CloseMicrophone() {
	speaker.Clear()
}

type opusReader struct {
	input       chan *[]byte
	opusDecoder opus.Decoder

	decodeBuffer       []byte
}

func (o *opusReader) Read(p []byte) (n int, err error) {
	segment := <-o.input
	if _, _, err = o.opusDecoder.Decode(*segment, o.decodeBuffer); err != nil {
		return 0,err
	}
	n = copy(p, o.decodeBuffer)
	return n, nil
}

// pcmStream allows faiface to play PCM directly
type pcmStream struct {
	reader   io.Reader
	format   beep.Format
	buf []byte
	len int
	pos int
	err error
}

func (s *pcmStream) Err() error { return s.err }

func (s *pcmStream) Stream(samples [][2]float64) (n int, ok bool) {
	width := s.format.Width()
	// if there's not enough data for a full sample, get more
	if size := s.len - s.pos; size < width {
		// if there's a partial sample, move it to the beginning of the buffer
		if size != 0 {
			copy(s.buf, s.buf[s.pos:s.len])
		}
		s.len = size
		s.pos = 0
		// refill the buffer
		nbytes, err := s.reader.Read(s.buf[s.len:])
		if err != nil {
			if err != io.EOF {
				s.err = err
			}
			return n, false
		}
		s.len += nbytes
	}
	// decode as many samples as we can
	for n < len(samples) && s.len-s.pos >= width {
		samples[n], _ = s.format.DecodeSigned(s.buf[s.pos:])
		n++
		s.pos += width
	}
	return n, true
}
