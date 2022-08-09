package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	HIDproxyEndpoint = "localhost:4000"
)

const (
	mouseWheel = 0
	mouseMove = 1
	mouseBtnUp = 2
	mouseBtnDown = 3
	
	keyUp = 4
	keyDown = 5
	keyPress = 6
	keyReset = 7
)


type HIDMsg struct {
	EventCode int			`json:"code"`
	Data map[string]float32 `json:"data"`
}

func ParseHIDInput(data string) {
	var err error;
	var route string;
	var out []byte;

	bodymap := make(map[string]float32)
	var bodystr float32
	bodystr = 0;

	var msg HIDMsg;
	json.Unmarshal([]byte(data),&msg);
	switch msg.EventCode {
	case mouseWheel:
		route = "Mouse/Wheel"
		bodystr = msg.Data["deltaY"];
	case mouseBtnUp:
		route = "Mouse/Up"
		bodystr = msg.Data["button"];
	case mouseBtnDown:
		route = "Mouse/Down"
		bodystr = msg.Data["button"];
	case mouseMove:
		route = "Mouse/Move"
		bodymap["X"] = msg.Data["dX"];
		bodymap["Y"] = msg.Data["dY"];

	case keyUp:
		route = "Keyboard/Up"
		bodystr = msg.Data["key"];
	case keyDown:
		route = "Keyboard/Down"
		bodystr = msg.Data["key"];

	case keyReset:
		route = "Keyboard/Reset"
	case keyPress:
		route = "Keyboard/Press"
	}

	
	if bodystr != 0 {
		out,err = json.Marshal(bodystr)
	} else if len(bodymap) != 0 {
		out,err = json.Marshal(bodymap)
	} else {
		out = []byte("");
	}

	if err != nil {
		fmt.Printf("fail to marshal output: %s\n",err.Error());
	}
	fmt.Printf("req: %s\n",string(out));
	res,err := http.Post(fmt.Sprintf("http://%s/%s",HIDproxyEndpoint,route),
		"application/json",bytes.NewBuffer(out));
	if err != nil {
		fmt.Printf("fail to forward input: %s\n",err.Error());
	}
	buf := make([]byte,100);
	res.Body.Read(buf);
	fmt.Printf("res: %s\n",string(buf));
}