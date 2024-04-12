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
	Type  string `json:"type"`
	Value int    `json:"value"`
}

func NewManualCtx(queue *proxy.Queue) datachannel.DatachannelConsumer {
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
				fmt.Printf("error unmarshal packet %s\n", err.Error())
				continue
			}

			switch dat.Type {
			case "bitrate":
				queue.Raise(proxy.Bitrate,dat.Value)
			case "framerate":
				queue.Raise(proxy.Framerate,dat.Value)
			case "pointer":
				queue.Raise(proxy.Pointer,dat.Value)
			case "reset":
				queue.Raise(proxy.Idr,dat.Value)
			case "danger-reset":
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
