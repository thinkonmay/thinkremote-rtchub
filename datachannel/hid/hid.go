package hid

import (
	"fmt"
	"strconv"
	"strings"
	"encoding/base64"

	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
)

const (
	HIDdefaultEndpoint = "localhost:5000"
)

var (
	queue_size = 100
	keys = []string{ 
		"Down" , "Up" , "Left" , "Right" , "Enter" , "Esc" , "Alt" , "Control" ,
        "Shift" , "PAUSE" , "BREAK" , "Backspace" , "Tab" , "CapsLock" , "Delete" , "Home" , "End" , "PageUp" ,
        "PageDown" , "NumLock" , "Insert" , "ScrollLock" , "F1" , "F2" , "F3" , "F4" , "F5" ,
        "F6" , "F7" , "F8" , "F9" , "F10" , "F11" , "F12" , "Meta",
	}
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
				Msg: fmt.Sprintf("%s|%s",v,vibration.SmallMotor,vibration.LargeMotor),
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


	go func() {
		for {
			message := <-ret.recv
			msg := strings.Split(message, "|")
			switch msg[0] {
			case "mma":
				x,_ := strconv.ParseFloat(msg[1],32)
				y,_ := strconv.ParseFloat(msg[2],32)
				go SendMouseAbsolute(float32(x),float32(y))
				continue;
			case "mmr":
				x,_ := strconv.ParseFloat(msg[1],32)
				y,_ := strconv.ParseFloat(msg[2],32)
				go SendMouseRelative(float32(x),float32(y))
				continue;
			case "mw":
				x,_ := strconv.ParseFloat(msg[1],32)
				go SendMouseWheel(x)
				continue;
			case "mu":
				x,_ := strconv.ParseInt(msg[1],10,8)
				go SendMouseButton(int(x),true)
				continue;
			case "md":
				x,_ := strconv.ParseInt(msg[1],10,8)
				go SendMouseButton(int(x),false)
				continue;
			case "ku":
				go SendKeyboard(msg[1],true)
				continue;
			case "kd":
				go SendKeyboard(msg[1],false)
				continue;
			case "kr":
				go func ()  { for _,v := range keys { SendKeyboard(v,true) } }() 
				continue;
			case "gs":
				x,_ := strconv.ParseInt(msg[2],10,32)
				y,_ := strconv.ParseFloat(msg[3],32)
                controller.pressSlider(x,y);
				continue;
            case "ga":
				x,_ := strconv.ParseInt(msg[2],10,32)
				y,_ := strconv.ParseFloat(msg[3],32)
                controller.pressAxis(x,y);
				continue;
            case "gb":
				y,_ := strconv.ParseInt(msg[2],10,32)
                controller.pressButton(y,msg[3] == "1");
				continue;
            case "cs":
				decoded,err := base64.RawStdEncoding.DecodeString(msg[1])
				if err != nil {
					fmt.Println(err.Error())
					continue;
				}

                SetClipboard(string(decoded));
                continue;
			}
		}
	}()


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