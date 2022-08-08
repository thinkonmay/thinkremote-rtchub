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
	Data map[string]string	`json:"data"`
}

func ParseHIDInput(data string) {
	var err error;
	var route string;
	var out []byte;

	bodymap := make(map[string]string)
	bodystr := "";

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

	
	if bodystr != "" {
		out = []byte(bodystr);
	} else if len(bodymap) != 0 {
		out,err = json.Marshal(bodymap)
	} else {
		out = []byte("");
	}

	if err != nil {
		fmt.Printf("fail to marshal output: %s",err.Error());
	}
	http.Post(fmt.Sprintf("http://%s/%s",HIDproxyEndpoint,route),
		"application/json",bytes.NewBuffer(out));
}