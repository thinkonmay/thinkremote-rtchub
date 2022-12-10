package iceservers

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/pion/webrtc/v3"
)

func TestFilter(t *testing.T) {
	rtc := webrtc.Configuration{ ICEServers: []webrtc.ICEServer{{
			URLs: []string{
				"stun:stun.l.google.com:19302",
			}, }, {
				URLs:           []string{"turn:workstation.thinkmay.net:3478"},
				Username:       "oneplay",
				Credential:     "oneplay",
				CredentialType: webrtc.ICECredentialTypePassword,
			}, {
				URLs:           []string{"turn:stun.l.google.com:19302"},
				Username:       "oneplay",
				Credential:     "oneplay",
				CredentialType: webrtc.ICECredentialTypePassword,
		}},
	}

	result := FilterWebRTCConfig(rtc)
	byt,_ := json.MarshalIndent(result," "," ")
	fmt.Printf("%s\n",string(byt));
}



