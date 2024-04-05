package manual

import (
	"encoding/json"
	"fmt"

	proxy "github.com/thinkonmay/thinkremote-rtchub"
	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
)

const (
	queue_size = 32
)

type Manual struct {
	In   chan string
	ctxs []string
	Out  chan struct {
		id  string
		val string
	}
}

type ManualPacket struct {
	Type  int `json:"type"`
	Value int `json:"value"`
}

func NewManualCtx(memory *proxy.SharedMemory) datachannel.DatachannelConsumer {
	ret := &Manual{
		In: make(chan string, queue_size),
		Out: make(chan struct {
			id  string
			val string
		}, queue_size),
		ctxs: []string{},
	}

	go func() {
		for {
			data := <-ret.In
			dat := ManualPacket{}
			err := json.Unmarshal([]byte(data), &dat)
			if err != nil {
				fmt.Printf("%s", err.Error())
				continue
			}

		}
	}()

	return ret
}

// TODO
func (ads *Manual) Recv() (string, string) {
	out := <-ads.Out
	return out.id, out.val
}
func (ads *Manual) Send(id string, msg string) {
	ads.In <- msg
}

func (ads *Manual) SetContext(ids []string) {
	ads.ctxs = ids
}
