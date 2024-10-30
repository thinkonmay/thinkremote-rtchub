package manual

import (
	"encoding/json"
	"fmt"

	proxy "github.com/thinkonmay/thinkremote-rtchub"
	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
	"github.com/thinkonmay/thinkremote-rtchub/util/thread"
)

const (
	queue_size = 32
)

type Manual struct {
	In  chan string
	Out chan string
}

type ManualPacket struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

func NewManualCtx(queue *proxy.Queue) datachannel.DatachannelConsumer {
	ret := &Manual{
		In:  make(chan string, queue_size),
		Out: make(chan string, queue_size),
	}

	dat := ManualPacket{}
	thread.SafeLoop(make(chan bool), 0, func() {
		if err := json.Unmarshal([]byte(<-ret.In), &dat); err != nil {
			fmt.Printf("error unmarshal packet %s\n", err.Error())
		} else {
			switch dat.Type {
			case "bitrate":
				queue.Raise(proxy.Bitrate, dat.Value)
			case "framerate":
				queue.Raise(proxy.Framerate, dat.Value)
			case "pointer":
				queue.Raise(proxy.Pointer, dat.Value)
			case "reset":
				queue.Raise(proxy.Idr, dat.Value)
			case "danger-reset":
			}
		}

	})

	return ret
}

// TODO
func (manual *Manual) Recv() chan string {
	return manual.Out
}
func (manual *Manual) Send(msg string) {
	manual.In <- msg
}
