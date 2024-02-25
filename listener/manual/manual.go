package manual

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
)

const (
	queue_size = 32
)

type Manual struct {
	In         chan string
	ctxs       []string
	Out        chan struct{
		id string
		val string
	} 


	triggerVideoReset func()
	bitrateCallback func(bitrate int)
	framerateCallback func(framerate int)
	pointerCallback func(pointer int)
	displayCallback func(display string)
}

func NewManualCtx( BitrateCallback func(bitrate int),
				   FramerateCallback func(framerate int),
				   PointerCallback  func(pointer int),
				   DisplayCallback  func(display string),
				   CodecCallback  func(codec string),
				   IDRcallback func()) datachannel.DatachannelConsumer {
	ret := &Manual{
		In:         make(chan string,queue_size),
		Out:        make(chan struct{id string;val string},queue_size),
		ctxs: 		[]string{},

		triggerVideoReset: IDRcallback,
		bitrateCallback: BitrateCallback,
		framerateCallback: FramerateCallback,
		pointerCallback: PointerCallback,
		displayCallback: DisplayCallback,
	}

	go func() { for { data := <-ret.In

			var dat map[string]interface{}
			err := json.Unmarshal([]byte(data), &dat)
			if err != nil {
				fmt.Printf("%s", err.Error())
				continue
			}

			if dat["type"] == nil {
				continue
			}
			_type := dat["type"].(string)
			if _type == "bitrate" && dat["value"] != nil{
				_val  := dat["value"].(float64)
				ret.bitrateCallback(int(_val))
			} else if _type == "framerate" && dat["value"] != nil{
				_val  := dat["value"].(float64)
				ret.framerateCallback(int(_val))
			} else if _type == "pointer" && dat["value"] != nil{
				_val  := dat["value"].(float64)
				ret.pointerCallback(int(_val))
			} else if _type == "display" && dat["value"] != nil{
				ret.displayCallback(dat["value"].(string))
			} else if _type == "reset" {
				ret.triggerVideoReset()
			} else if _type == "danger-reset" {
				os.Exit(0)
			}
		}
	}()



	return ret
}


// TODO
func (ads *Manual)Recv() (string,string) {
	out := <-ads.Out
	return out.id,out.val
}
func (ads *Manual)Send(id string,msg string) {
	ads.In<-msg
}

func (ads *Manual)SetContext(ids []string) {
	ads.ctxs = ids
}