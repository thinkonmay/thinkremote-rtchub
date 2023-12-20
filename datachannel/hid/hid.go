package hid

import (
	"fmt"
	"strconv"
	"strings"
	"encoding/base64"

	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
	"github.com/thinkonmay/thinkremote-rtchub/util/win32"
)

const (
	HIDdefaultEndpoint = "localhost:5000"
)

var (
	queue_size = 128
)

type HIDAdapter struct {
	send chan datachannel.Msg
	recv chan string

	ids []string
}

func NewHIDSingleton() datachannel.DatachannelConsumer {
	ret := HIDAdapter{
		send: make(chan datachannel.Msg,queue_size),
		recv: make(chan string,queue_size),
		ids: []string{},
	}

	em,err := NewEmulator(func(vibration Vibration) {
		for _,v := range ret.ids {
			ret.send <- datachannel.Msg{
				Msg: fmt.Sprintf("%d|%d",int(vibration.SmallMotor),int(vibration.LargeMotor)),
				Id: v,
			}
		}
	})
	if err != nil {
		fmt.Printf("%s\n",err.Error())
	}

	controller,err := em.CreateXbox360Controller()
	if err != nil {
		fmt.Printf("%s\n",err.Error())
	}

	err = controller.Connect()
	if err != nil {
		fmt.Printf("%s\n",err.Error())
	}


	process := func() { win32.HighPriorityThread()
		for { message := <-ret.recv
			msg := strings.Split(message, "|")
			switch msg[0] {
			case "mma":
				x,_ := strconv.ParseFloat(msg[1],32)
				y,_ := strconv.ParseFloat(msg[2],32)
				SendMouseAbsolute(float32(x),float32(y))
			case "mmr":
				x,_ := strconv.ParseFloat(msg[1],32)
				y,_ := strconv.ParseFloat(msg[2],32)
				SendMouseRelative(float32(x),float32(y))
			case "mw":
				x,_ := strconv.ParseFloat(msg[1],32)
				SendMouseWheel(x)
			case "mu":
				x,_ := strconv.ParseInt(msg[1],10,8)
				SendMouseButton(int(x),true)
			case "md":
				x,_ := strconv.ParseInt(msg[1],10,8)
				SendMouseButton(int(x),false)
			case "ku":
				x,_ := strconv.ParseInt(msg[1],10,32)
				SendKeyboard(int(x),true,false)
			case "kd":
				x,_ := strconv.ParseInt(msg[1],10,32)
				SendKeyboard(int(x),false,false)
			case "kus":
				x,_ := strconv.ParseInt(msg[1],10,32)
				SendKeyboard(int(x),true,true)
			case "kds":
				x,_ := strconv.ParseInt(msg[1],10,32)
				SendKeyboard(int(x),false,true)
			case "kr":
				for i := 0; i < 0xFF ; i++ { 
					SendKeyboard(i,true,false) 
				} 
			case "gs":
				x,_ := strconv.ParseInt(msg[2],10,32)
				y,_ := strconv.ParseFloat(msg[3],32)
                controller.pressSlider(x,y);
            case "ga":
				x,_ := strconv.ParseInt(msg[2],10,32)
				y,_ := strconv.ParseFloat(msg[3],32)
                controller.pressAxis(x,y);
            case "gb":
				y,_ := strconv.ParseInt(msg[2],10,32)
                controller.pressButton(y,msg[3] == "1");
            case "cs":
				decoded,err := base64.StdEncoding.DecodeString(msg[1])
				if err != nil {
					fmt.Println(err.Error())
					continue;
				}

                SetClipboard(string(decoded));
			}
		}
	}


	go process()
	return &ret
}

func (hid *HIDAdapter) Recv() (string,string) {
	out := <-hid.send
	return out.Id,out.Msg

}
func (hid *HIDAdapter) Send(id string,msg string) {
	hid.recv<-fmt.Sprintf("%s|%s",msg,id)
}

func (hid *HIDAdapter) SetContext(ids []string) {
	hid.ids = ids
}