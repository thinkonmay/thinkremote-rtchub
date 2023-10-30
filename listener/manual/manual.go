package manual

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
	"github.com/thinkonmay/thinkremote-rtchub/listener/display"
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
	pointerCallback func(pointer int)
	displayCallback func(display map[string]interface{})
}

func NewManualCtx(BitrateCallback func(bitrate int),
				   PointerCallback  func(pointer int),
				   DisplayCallback  func(display map[string]interface{}),
				   CodecCallback  func(codec string),
				   IDRcallback func()) datachannel.DatachannelConsumer {
	ret := &Manual{
		In:         make(chan string,100),
		Out:        make(chan struct{id string;val string},100),
		ctxs: 		[]string{},

		triggerVideoReset: IDRcallback,
		bitrateCallback: BitrateCallback,
		pointerCallback: PointerCallback,
		displayCallback: DisplayCallback,
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
			} else if _type == "pointer" && dat["value"] != nil{
				_val  := dat["value"].(float64)
				ret.pointerCallback(int(_val))
			} else if _type == "displays" {
				b,_ := json.Marshal(map[string]interface{}{
					"type": "displays",
					"value": display.GetDisplays(),
				})
				for _,v := range ret.ctxs {
				ret.Out<-struct{id string; val string}{
					id : v,
					val: string(b),
				}}
			} else if _type == "display" && dat["value"] != nil{
				ret.displayCallback(dat["value"].(map[string]interface{}))
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