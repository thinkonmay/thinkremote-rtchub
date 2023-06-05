package hid

import (
	"context"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
)

const (
	HIDdefaultEndpoint = "localhost:5000"
)

const (
	queue_size = 100
)

type HIDAdapter struct {
	client      *websocket.Conn
	URL         string

	send chan string
	recv chan string
}

func NewHIDSingleton(URL string) datachannel.DatachannelConsumer {
	if URL == "" {
		panic("no URL provided")
	}

	ret := HIDAdapter{
		URL:         URL,
		send: make(chan string,queue_size),
		recv: make(chan string,queue_size),
	}

	setup := func () bool {
		var err error
		dialer := websocket.Dialer{ HandshakeTimeout: time.Second, }
		dial_ctx,_ := context.WithTimeout(context.TODO(),time.Second)
		ret.client, _, err = dialer.DialContext(dial_ctx,fmt.Sprintf("ws://%s/Socket", URL), nil)
		if err != nil || ret.client == nil {
			fmt.Printf("hid websocket error: %s", err.Error())
			return false
		}

		err = ret.client.WriteMessage(websocket.TextMessage, []byte("ping"))
		if err != nil {
			fmt.Printf("hid websocket error: %s", err.Error())
			return false
		}

		return true
	}
	go func() {
		for {
			if ret.client == nil {
				if setup(){
					continue
				}

				fmt.Println("fail to setup hid adapter")
			}

			time.Sleep(time.Millisecond * 1000)
		}
	}()

	go func() {
		for {
			message := <-ret.recv
			if ret.client == nil {
				continue
			}

			if ret.client.WriteMessage(websocket.TextMessage, []byte(message)) != nil {
				ret.client = nil
			}
		}
	}()

	go func() {
		for {
			if ret.client == nil {
				time.Sleep(time.Millisecond * 100)
				continue
			}

			typ, message, err := ret.client.ReadMessage()
			if err != nil || typ == websocket.CloseMessage {
				ret.client = nil
			}

			if typ != websocket.TextMessage {
				continue
			}

			ret.send <- string(message)
		}
	}()

	return &ret
}

func (hid *HIDAdapter) Recv() string {
	return <-hid.send

}
func (hid *HIDAdapter) Send(msg string) {
	hid.recv<-msg
}