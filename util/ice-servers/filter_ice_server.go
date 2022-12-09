package iceservers

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-ping/ping"
	"github.com/pion/webrtc/v3"
)

func FilterWebRTCConfig(config webrtc.Configuration) (webrtc.Configuration){
	result := webrtc.Configuration{}

	total_turn,count := 0,0
	pingResults := map[string]time.Duration{}
	for _,ice := range config.ICEServers {
		splits := strings.Split(ice.URLs[0],":");
		if splits[0] == "turn" {
			total_turn++;
			go func ()  {
				defer func ()  {
					count++;
				}()

				domain := splits[1];
				pinger, err := ping.NewPinger(domain)
				pinger.SetPrivileged(true)
				if err != nil {
					return
				}
				pinger.Count = 5
				err = pinger.Run() // Blocks until finished.
				if err != nil {
					return 
				}

				stats := pinger.Statistics() // get send/receive/duplicate/rtt stats
				fmt.Printf("stats %s %d\n",ice.URLs[0],stats.AvgRtt.Milliseconds());

				pingResults[ice.URLs[0]] = stats.AvgRtt
			}()
		} else if splits[0] == "stun" {
			result.ICEServers = append(result.ICEServers, ice)
			continue
		}
	}

	for {
		time.Sleep(100 * time.Millisecond)
		if count == total_turn {
			break
		}
	}

	minUrl,minDuration := "", 100 *time.Second
	for url,result := range pingResults {
		if result < minDuration {
			minDuration = result
			minUrl = url
		}
	}

	for _,ice := range config.ICEServers {
		if ice.URLs[0] == minUrl {
			result.ICEServers = append(result.ICEServers, ice)
		}
	}


	return result
}