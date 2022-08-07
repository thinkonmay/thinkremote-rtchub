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
)


type HIDMsg struct {
	EventCode int			`json:"code"`
	Data map[string]int		`json:"data"`

}

func ParseHIDInput(data string) {
	var route string;
	var body map[string]string;

	var msg HIDMsg;
	json.Unmarshal([]byte(data),&msg);
	switch msg.EventCode {
	case mouseWheel:
	case mouseBtnUp:
	case mouseBtnDown:
	case mouseMove:

	case keyDown:
	case keyUp:
	}

	
	out,err := json.Marshal(body);
	if err != nil {
		fmt.Printf("fail to marshal output");
	}
	http.Post(fmt.Sprintf("http://%s",HIDproxyEndpoint),"application/json",bytes.NewBuffer(out));
}