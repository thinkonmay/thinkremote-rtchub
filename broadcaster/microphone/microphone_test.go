package microphone

import (
	"testing"
	"time"

	"github.com/pion/rtp"
	"github.com/thinkonmay/thinkremote-rtchub/listener/audio"
)

const (
	audio_default = "wasapisrc name=source device=\"\\{0.0.1.00000000\\}.\\{1b950755-f779-48bc-9813-674059b82cfd\\}\" ! audioresample ! audio/x-raw,rate=48000 ! audioconvert ! opusenc name=encoder ! appsink name=appsink"
)
func TestMic(t *testing.T) {
	audiopipeline,err := audio.CreatePipeline(audio_default)
	if err != nil {
		t.Error(err.Error())
		return
	}
	mic,err := CreatePipeline(96)
	if err != nil {
		t.Error(err.Error())
		return
	}

	audiopipeline.Open()
	mic.Open()
	defer audiopipeline.Close()
	defer mic.Close()
	audiopipeline.RegisterRTPHandler("0",func(pkt *rtp.Packet) {
		buff,err := pkt.Marshal()
		if err != nil {
			t.Error(err.Error())
			return
		}
		
		mic.Push(buff)
	})

	time.Sleep(2000 * time.Second)
}