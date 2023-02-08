package gsttest

import (
	"fmt"
	"testing"

	"github.com/thinkonmay/thinkremote-rtchub/util/tool"
)

func TestTest(t *testing.T) {
	dev := tool.GetDevice()
	result := GstTestVideo(&dev.Monitors[0])
	fmt.Printf("%s\n", result)

	souncard := tool.Soundcard{}
	for _, card := range dev.Soundcards {
		if card.Name == "Default Audio Render Device" {
			souncard = card

		}
	}

	result = GstTestAudio(&souncard)
	fmt.Printf("%s\n", result)
}
