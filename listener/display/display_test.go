package display

import (
	"fmt"
	"testing"
)

func TestDisplay(t *testing.T) {
	fmt.Printf("%v\n",GetDisplays()[0])
	SetResolution("\\\\.\\DISPLAY2",1920,1200)
}