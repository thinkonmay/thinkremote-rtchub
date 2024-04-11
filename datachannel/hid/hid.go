package hid

/*
#include "Input.h"
*/
import "C"
import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"unsafe"

	proxy "github.com/thinkonmay/thinkremote-rtchub"
	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
	"github.com/thinkonmay/thinkremote-rtchub/util/thread"
)

const (
	queue_size                 = 128
	SS_KBE_FLAG_NON_NORMALIZED = 1
)

type HIDAdapter struct {
	send chan datachannel.Msg
	recv chan string

	ids []string
}

func NewHIDSingleton(queue *proxy.Queue) datachannel.DatachannelConsumer {
	ret := HIDAdapter{
		send: make(chan datachannel.Msg, queue_size),
		recv: make(chan string, queue_size),
		ids:  []string{},
	}

	process := func() {
		thread.HighPriorityThread()
		for {
			message := <-ret.recv
			msg := strings.Split(message, "|")
			switch msg[0] {
			case "mma":
				x, _ := strconv.ParseFloat(msg[1], 32)
				y, _ := strconv.ParseFloat(msg[2], 32)
				packet := C.NV_ABS_MOUSE_MOVE_PACKET{
					header: C.NV_INPUT_HEADER{
						magic: C.MOUSE_MOVE_ABS_MAGIC,
					},
					x:      C.short(x),
					y:      C.short(y),
				}
				queue.Write(packet, int(unsafe.Sizeof(packet)))
			case "mmr":
				x, _ := strconv.ParseFloat(msg[1], 32)
				y, _ := strconv.ParseFloat(msg[2], 32)
				packet := C.NV_REL_MOUSE_MOVE_PACKET{
					header: C.NV_INPUT_HEADER{
						magic: C.MOUSE_MOVE_REL_MAGIC,
					},
					deltaX: C.short(x),
					deltaY: C.short(y),
				}
				queue.Write(packet, int(unsafe.Sizeof(packet)))
			case "mw":
				x, _ := strconv.ParseFloat(msg[1], 32)
				packet := C.SS_HSCROLL_PACKET{
					header:     C.NV_INPUT_HEADER{
						magic: C.SCROLL_MAGIC_GEN5,
					},
					scrollAmount: C.short(x),
				}
				queue.Write(packet, int(unsafe.Sizeof(packet)))
			case "mu":
				x, _ := strconv.ParseInt(msg[1], 10, 8)
				packet := C.NV_MOUSE_BUTTON_PACKET{
					header: C.NV_INPUT_HEADER{
						magic: C.MOUSE_BUTTON_UP_EVENT_MAGIC_GEN5,
					},
					button: C.uchar(x),
				}
				queue.Write(packet, int(unsafe.Sizeof(packet)))
			case "md":
				x, _ := strconv.ParseInt(msg[1], 10, 8)
				packet := C.NV_MOUSE_BUTTON_PACKET{
					header: C.NV_INPUT_HEADER{
						magic: C.MOUSE_BUTTON_DOWN_EVENT_MAGIC_GEN5,
					},
					button: C.uchar(x),
				}
				queue.Write(packet, int(unsafe.Sizeof(packet)))
			case "ku":
				packet := C.NV_KEYBOARD_PACKET{
					header: C.NV_INPUT_HEADER{
						magic: C.KEY_UP_EVENT_MAGIC,
					},
					flags: C.char(SS_KBE_FLAG_NON_NORMALIZED),
				}
				queue.Write(packet, int(unsafe.Sizeof(packet)))
			case "kd":
				packet := C.NV_KEYBOARD_PACKET{
					header: C.NV_INPUT_HEADER{
						magic: C.KEY_DOWN_EVENT_MAGIC,
					},
					flags: C.char(SS_KBE_FLAG_NON_NORMALIZED),
				}
				queue.Write(packet, int(unsafe.Sizeof(packet)))
			case "kus":
				packet := C.NV_KEYBOARD_PACKET{
					header: C.NV_INPUT_HEADER{
						magic: C.KEY_UP_EVENT_MAGIC,
					},
				}
				queue.Write(packet, int(unsafe.Sizeof(packet)))
			case "kds":
				packet := C.NV_KEYBOARD_PACKET{
					header: C.NV_INPUT_HEADER{
						magic: C.KEY_DOWN_EVENT_MAGIC,
					},


				}
				queue.Write(packet, int(unsafe.Sizeof(packet)))
			case "kr":
			case "gs":
				packet := C.NV_MULTI_CONTROLLER_PACKET{
					header: C.NV_INPUT_HEADER{},
				}
				queue.Write(packet, int(unsafe.Sizeof(packet)))
			case "ga":
				packet := C.NV_MULTI_CONTROLLER_PACKET{
					header: C.NV_INPUT_HEADER{},
				}
				queue.Write(packet, int(unsafe.Sizeof(packet)))
			case "gb":
				packet := C.NV_MULTI_CONTROLLER_PACKET{
					header: C.NV_INPUT_HEADER{},
				}
				queue.Write(packet, int(unsafe.Sizeof(packet)))
			case "cs":
				decoded, err := base64.StdEncoding.DecodeString(msg[1])
				if err != nil {
					fmt.Println(err.Error())
					continue
				}

				SetClipboard(string(decoded))
			}
		}
	}

	go process()
	return &ret
}

func SetClipboard(s string) {
}

func (hid *HIDAdapter) Recv() (string, string) {
	out := <-hid.send
	return out.Id, out.Msg

}
func (hid *HIDAdapter) Send(id string, msg string) {
	hid.recv <- fmt.Sprintf("%s|%s", msg, id)
}

func (hid *HIDAdapter) SetContext(ids []string) {
	hid.ids = ids
}
