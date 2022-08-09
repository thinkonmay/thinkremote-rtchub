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
	Data map[string]interface{} `json:"data"`
}

func ParseHIDInput(data string) {
	var err error;
	var route string;
	var out []byte;

	bodymap := make(map[string]float64)
	var bodyfloat float64
	var bodystring string 
	bodyfloat = -1;
	bodystring = "";

	var msg HIDMsg;
	json.Unmarshal([]byte(data),&msg);
	json.Unmarshal([]byte(data),&msg);
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
	}

	
	if bodyfloat != -1 {
		out,err = json.Marshal(bodyfloat)
	} else if bodystring != "" {
		out = []byte(bodystring)
	} else if len(bodymap) != 0 {
		out,err = json.Marshal(bodymap)
	} else {
		out = []byte("");
	}

	if err != nil {
		fmt.Printf("fail to marshal output: %s\n",err.Error());
	}
	fmt.Printf("req: %s\n",string(out));
	_,err = http.Post(fmt.Sprintf("http://%s/%s",HIDproxyEndpoint,route),
		"application/json",bytes.NewBuffer(out));
	if err != nil {
		fmt.Printf("fail to forward input: %s\n",err.Error());
	}
}