package manual

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
)


type Manual struct {
	In         chan string
	Out        chan string

	triggerVideoReset func()
	audioResetCallback func()
	bitrateCallback func(bitrate int)
	framerateCallback func(framerate int)
	pointerCallback func(pointer int)
}

func NewManualCtx(BitrateCallback func(bitrate int),
				   FramerateCallback func(framerate int),
				   PointerCallback  func(pointer int),
				   IDRcallback func(),
				   AudioResetcallback func()) datachannel.DatachannelConsumer {
	ret := &Manual{
		In:         make(chan string,100),
		Out:        make(chan string,100),

		audioResetCallback: AudioResetcallback,
		triggerVideoReset: IDRcallback,
		bitrateCallback: BitrateCallback,
		framerateCallback: FramerateCallback,
		pointerCallback: PointerCallback,
	}

	go func() {
		for {
			data := <-ret.In

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
			} else if _type == "reset" {
				ret.triggerVideoReset()
			} else if _type == "audio-reset" {
				ret.audioResetCallback()
			} else if _type == "danger-reset" {
				os.Exit(0)
			}
		}
	}()



	return ret
}


// TODO
func (ads *Manual)Recv() (string,string) {
	<-ads.Out
	return "",""
}
func (ads *Manual)Send(id string,msg string) {
	ads.In<-msg
}

func (ads *Manual)SetContext(ids []string) {
}