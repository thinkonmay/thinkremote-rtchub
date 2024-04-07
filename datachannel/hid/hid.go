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

func NewHIDSingleton(memory *proxy.SharedMemory) datachannel.DatachannelConsumer {
	ret := HIDAdapter{
		send: make(chan datachannel.Msg, queue_size),
		recv: make(chan string, queue_size),
		ids:  []string{},
	}

	_ = memory.GetQueue(proxy.Input)
	process := func() {
		thread.HighPriorityThread()
		for {
			message := <-ret.recv
			msg := strings.Split(message, "|")
			switch msg[0] {
			case "mma":
				x, _ := strconv.ParseFloat(msg[1], 32)
				y, _ := strconv.ParseFloat(msg[2], 32)
				_ = C.NV_ABS_MOUSE_MOVE_PACKET{
					header: C.NV_INPUT_HEADER{},
					x:      C.short(x),
					y:      C.short(y),
				}
			case "mmr":
				x, _ := strconv.ParseFloat(msg[1], 32)
				y, _ := strconv.ParseFloat(msg[2], 32)
				_ = C.NV_REL_MOUSE_MOVE_PACKET{
					header: C.NV_INPUT_HEADER{},
					deltaX: C.short(x),
					deltaY: C.short(y),
				}
			case "mw":
				x, _ := strconv.ParseFloat(msg[1], 32)
				_ = C.NV_SCROLL_PACKET{
					header:     C.NV_INPUT_HEADER{},
					scrollAmt1: C.short(x),
				}
			case "mu":
				x, _ := strconv.ParseInt(msg[1], 10, 8)
				_ = C.NV_MOUSE_BUTTON_PACKET{
					header: C.NV_INPUT_HEADER{
						magic: C.MOUSE_BUTTON_UP_EVENT_MAGIC_GEN5,
					},
					button: C.uchar(x),
				}
			case "md":
				x, _ := strconv.ParseInt(msg[1], 10, 8)
				_ = C.NV_MOUSE_BUTTON_PACKET{
					header: C.NV_INPUT_HEADER{
						magic: C.MOUSE_BUTTON_DOWN_EVENT_MAGIC_GEN5,
					},
					button: C.uchar(x),
				}
			case "ku":
				_ = C.NV_KEYBOARD_PACKET{
					header: C.NV_INPUT_HEADER{
						magic: C.KEY_UP_EVENT_MAGIC,
					},
				}
			case "kd":
				_ = C.NV_KEYBOARD_PACKET{
					header: C.NV_INPUT_HEADER{
						magic: C.KEY_DOWN_EVENT_MAGIC,
					},
				}
			case "kus":
				_ = C.NV_KEYBOARD_PACKET{
					header: C.NV_INPUT_HEADER{
						magic: C.KEY_UP_EVENT_MAGIC,
					},
					flags: C.char(SS_KBE_FLAG_NON_NORMALIZED),
				}
			case "kds":
				_ = C.NV_KEYBOARD_PACKET{
					header: C.NV_INPUT_HEADER{
						magic: C.KEY_DOWN_EVENT_MAGIC,
					},
					flags: C.char(SS_KBE_FLAG_NON_NORMALIZED),
				}
			case "kr":
			case "gs":
				_ = C.NV_MULTI_CONTROLLER_PACKET{
					header: C.NV_INPUT_HEADER{},
				}
			case "ga":
				_ = C.NV_MULTI_CONTROLLER_PACKET{
					header: C.NV_INPUT_HEADER{},
				}
			case "gb":
				_ = C.NV_MULTI_CONTROLLER_PACKET{
					header: C.NV_INPUT_HEADER{},
				}
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
	panic("unimplemented")
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
