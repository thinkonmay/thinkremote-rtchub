package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pion/webrtc/v3"
	proxy "github.com/thinkonmay/thinkremote-rtchub"
	"github.com/thinkonmay/thinkremote-rtchub/datachannel"
	"github.com/thinkonmay/thinkremote-rtchub/listener"
	"github.com/thinkonmay/thinkremote-rtchub/listener/audio"
	"github.com/thinkonmay/thinkremote-rtchub/signalling/websocket"
	"github.com/thinkonmay/thinkremote-rtchub/util/config"
)


const (
	url = "http://localhost:60000/handshake/server?token=audio"
)

func main() {
	args := os.Args[1:]
	webrtcArg := ""
	for i, arg := range args {
		if arg == "--webrtc" {
			webrtcArg = args[i+1]
		}
	}


	audioPipeline, err := audio.CreatePipeline()
	if err != nil {
		fmt.Printf("error initiate audio pipeline %s\n", err.Error())
		return
	}

	audioPipeline.Open()




	bytes1, _ := base64.StdEncoding.DecodeString(webrtcArg)
	var data map[string]interface{}
	json.Unmarshal(bytes1, &data)
	rtc := &config.WebRTCConfig{Ices: make([]webrtc.ICEServer, 0)}
	for _, v := range data["iceServers"].([]interface{}) {
		ice := webrtc.ICEServer{
			URLs: []string{v.(map[string]interface{})["urls"].(string)},
		}
		if v.(map[string]interface{})["credential"] != nil {
			ice.Credential = v.(map[string]interface{})["credential"].(string)
			ice.Username = v.(map[string]interface{})["username"].(string)
		}
		rtc.Ices = append(rtc.Ices, ice)
	}


	go func() { for {
			signaling_client, err := websocket.InitWebsocketClient(url)
			if err != nil {
				fmt.Printf("error initiate signaling client %s\n", err.Error())
				continue
			}

			_, err = proxy.InitWebRTCProxy(signaling_client, 
											rtc, 
											datachannel.NewDatachannel(), 
											[]listener.Listener{audioPipeline}, 
											func(tr *webrtc.TrackRemote) {})
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				continue
			}
			signaling_client.WaitForStart()
		}
	}()

	chann := make(chan os.Signal, 16)
	signal.Notify(chann, syscall.SIGTERM, os.Interrupt)
	<-chann
}
