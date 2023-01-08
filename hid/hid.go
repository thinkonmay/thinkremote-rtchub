package hid

import (
	"fmt"
	"time"

	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	"github.com/gorilla/websocket"
)

const (
	HIDdefaultEndpoint = "localhost:5000"
)

type HIDSingleton struct {
	datachaneel *config.DataChannel
	client *websocket.Conn
	URL string

	chann chan string
}


func NewHIDSingleton(URL string, DataChannel *config.DataChannel) *HIDSingleton{
	var err error
	ret := HIDSingleton{
		URL: URL,
		datachaneel: DataChannel,
	}

	if ret.URL == "" {
		ret.URL = HIDdefaultEndpoint	
	}

	ret.chann = make(chan string,100);
	go func() {
		for {
			ret.chann<-<-DataChannel.Recv
		}
	}()
	go func() {
		for {
			ret.chann<-"ping"
			time.Sleep(1 * time.Second)
		}
	}()
	go func() {
		for {
			message := <-ret.chann
			if ret.client != nil {
				err := ret.client.WriteMessage(websocket.TextMessage,[]byte(message));
				if err != nil {
					ret.client = nil
				}
			}
		}
	}()




	go func() {
		for {
			if ret.client == nil {
				ret.client, _, err = websocket.DefaultDialer.Dial(fmt.Sprintf("ws://%s/Socket",URL),nil)
				if err != nil || ret.client == nil{
					fmt.Println("hid websocket error: %s",err.Error())
					time.Sleep(time.Second)
					continue;
				}
				err := ret.client.WriteMessage(websocket.TextMessage,[]byte("ping"));
				if err != nil {
					fmt.Println("hid websocket error: %s",err.Error())
					time.Sleep(time.Second)
					continue;
				}
			}

			typ, message, err := ret.client.ReadMessage()
			if err != nil || typ == websocket.CloseMessage {
				ret.client.Close()
				ret.client = nil;
			}

			if typ ==  websocket.TextMessage {
				ret.datachaneel.Send<-string(message)
			}
		}
	}()

	return &ret;
}




