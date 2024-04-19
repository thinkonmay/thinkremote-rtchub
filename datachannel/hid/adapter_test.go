package hid

import (
	"fmt"
	"testing"
)

func TestSendKeyboard(t *testing.T){
	fmt.Println("Start Test")

	// SendKeyboard(KEY_Y, true, false) /* test with KEY_Y windows */
	// SendKeyboard(NUMPAD1, true, false) /* test with NUMPAD1 windows */
	// SendKeyboard(KEY_L, true, false) /* test with KEY_L windows */
	// SendKeyboard(KEY_A, true, false) /* test with KEY_Y windows */

	// Test Scancode
	SendKeyboard(NUMPAD1, true, true) /* test with KEY_A windows */

	fmt.Println("End Test")
}
