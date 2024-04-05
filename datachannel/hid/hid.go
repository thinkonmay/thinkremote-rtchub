package hid

import (
	"fmt"

	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
	"github.com/thinkonmay/thinkremote-rtchub/util/thread"
	"github.com/thinkonmay/thinkremote-rtchub"
)

const (
	queue_size = 128
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

	process := func() {
		thread.HighPriorityThread()
		for {
			_ = <-ret.recv

		}
	}

	go process()
	return &ret
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
