package hid

/*
#include "Input.h"
*/
import "C"
import (
	"encoding/base64"
	"fmt"
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
				// x, _ := strconv.ParseFloat(msg[1], 32)
				// y, _ := strconv.ParseFloat(msg[2], 32)
			case "mmr":
				// x, _ := strconv.ParseFloat(msg[1], 32)
				// y, _ := strconv.ParseFloat(msg[2], 32)
			case "mw":
				// x, _ := strconv.ParseFloat(msg[1], 32)
			case "mu":
				// x, _ := strconv.ParseInt(msg[1], 10, 8)
			case "md":
				// x, _ := strconv.ParseInt(msg[1], 10, 8)
			case "ku":
			case "kd":
			case "kus":
			case "kds":
			case "kr":
			case "gs":
			case "ga":
			case "gb":
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
