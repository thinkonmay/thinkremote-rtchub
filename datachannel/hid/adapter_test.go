package hid

import (
	"math/rand"
	"testing"
	"time"
)

func TestSendKeyboard(t *testing.T) {
	for i := float32(0); i < 100; i++ {
		SendKeyboard(KEY_A, false, false) /* test with KEY_A windows */
		SendKeyboard(KEY_A, true, false) /* test with KEY_A windows */
		time.Sleep(time.Millisecond * 100)
	}
}

func TestMoveMouse(t *testing.T) {
	for i := float32(0); i < 100; i++ {
		x := rand.Float32()*100 - 50
		y := rand.Float32()*100 - 50
		SendMouseRelative(x, y)
		time.Sleep(time.Millisecond * 100)
	}
	for i := float32(0); i < 100; i++ {
		x := rand.Float32()
		y := rand.Float32()
		SendMouseAbsolute(x, y)
		time.Sleep(time.Millisecond * 100)
	}
}

func TestButtonMouse(t *testing.T) {
	// SendMouseButton(0x02, true)
}
