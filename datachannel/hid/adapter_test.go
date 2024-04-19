package hid

import (
	"fmt"
	"testing"
)

func TestSendKeyboard(t *testing.T){
	fmt.Println("Start Test")

	SendKeyboard(0x30, true, false)

	fmt.Println("End Test")
}
