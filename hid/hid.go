package hid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

const (
	HIDdefaultEndpoint = "localhost:5000"

	mouseWheel = 0
	mouseMove = 1
	mouseBtnUp = 2
	mouseBtnDown = 3
	
	keyUp = 4
	keyDown = 5
	keyPress = 6
	keyReset = 7

    RelativeMouseOff = 8
    RelativeMouseOn = 9
)

type HIDSingleton struct {
	channel chan *HIDMsg
	client *http.Client
	URL string
}

type HIDMsg struct {
	EventCode int			`json:"code"`
	Data map[string]interface{} `json:"data"`
}


func NewHIDSingleton(URL string) *HIDSingleton{
	ret := HIDSingleton{
		URL: URL,
		channel: make(chan *HIDMsg,100),
		client: &http.Client{
			Timeout:   time.Second,
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: time.Second,
				}).Dial,
				TLSHandshakeTimeout: time.Second,
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}

	if ret.URL == "" {
		ret.URL = HIDdefaultEndpoint	
	}

	process := func ()  {
		var err error;
		var route string;
		var out []byte;


		for {
			bodyfloat := -1.0;
			bodystring := "";
			bodymap := make(map[string]float64)

			msg :=<-ret.channel;
			switch msg.EventCode {
			case mouseWheel:
				route = "Mouse/Wheel"
				bodyfloat = msg.Data["deltaY"].(float64);
			case mouseBtnUp:
				route = "Mouse/Up"
				bodyfloat = msg.Data["button"].(float64);
			case mouseBtnDown:
				route = "Mouse/Down"
				bodyfloat = msg.Data["button"].(float64);
			case mouseMove:
				route = "Mouse/Move"
				bodymap["X"] = msg.Data["dX"].(float64);
				bodymap["Y"] = msg.Data["dY"].(float64);
			case keyUp:
				route = "Keyboard/Up"
				bodystring = msg.Data["key"].(string);
			case keyDown:
				route = "Keyboard/Down"
				bodystring = msg.Data["key"].(string);
			case keyReset:
				route = "Keyboard/Reset"
			case keyPress:
				route = "Keyboard/Press"
			case RelativeMouseOff:
				route = "Mouse/Relative/Off"
			case RelativeMouseOn:
				route = "Mouse/Relative/On"
			}

			
			if bodyfloat != -1 {
				out,err = json.Marshal(bodyfloat)
			} else if bodystring != "" {
				out,err = json.Marshal(bodystring)
			} else if len(bodymap) != 0 {
				out,err = json.Marshal(bodymap)
			} else {
				out = []byte("");
			}

			if err != nil { continue; }
			go http.Post(fmt.Sprintf("http://%s/%s",ret.URL,route),
				"application/json",bytes.NewBuffer(out));
		}	
	};

	go process();
	return &ret;
}

func (hid *HIDSingleton)ParseHIDInput(data string) {
	var msg HIDMsg
	json.Unmarshal([]byte(data),&msg);
	hid.channel <- &msg;
}