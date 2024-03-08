package hid

import (
	"fmt"
	"testing"
)

func TestHID(t *testing.T) {
	x, y, width, height := DisplayPosition("\\\\.\\DISPLAY1")
	vx, vy := GetVirtualDisplay()

	fmt.Printf("%d %d %d %d %d %d", x, y, width, height, vx, vy)
}