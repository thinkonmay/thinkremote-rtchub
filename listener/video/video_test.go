package video

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/pion/rtp"
	"github.com/thinkonmay/thinkremote-rtchub/util/config"
)

func TestVideo(t *testing.T) {
	chans := config.NewDataChannelConfig([]string{"hid", "adaptive", "manual"})

	videoPipelineString := ""
	bytes1,_ := base64.StdEncoding.DecodeString("ImQzZDExc2NyZWVuY2FwdHVyZXNyYyBibG9ja3NpemU9ODE5MiBkby10aW1lc3RhbXA9dHJ1ZSBtb25pdG9yLWhhbmRsZT02NTUzNyAhIGNhcHNmaWx0ZXIgbmFtZT1mcmFtZXJhdGVGaWx0ZXIgISB2aWRlby94LXJhdyhtZW1vcnk6RDNEMTFNZW1vcnkpLGNsb2NrLXJhdGU9OTAwMDAgISBxdWV1ZSBtYXgtc2l6ZS10aW1lPTAgbWF4LXNpemUtYnl0ZXM9MCBtYXgtc2l6ZS1idWZmZXJzPTMgISBkM2QxMWNvbnZlcnQgISBxdWV1ZSBtYXgtc2l6ZS10aW1lPTAgbWF4LXNpemUtYnl0ZXM9MCBtYXgtc2l6ZS1idWZmZXJzPTMgISBxc3ZoMjY0ZW5jIGJpdHJhdGU9NjAwMCByYXRlLWNvbnRyb2w9MSBnb3Atc2l6ZT0tMSByZWYtZnJhbWVzPTEgbG93LWxhdGVuY3k9dHJ1ZSB0YXJnZXQtdXNhZ2U9NyBuYW1lPWVuY29kZXIgISB2aWRlby94LWgyNjQsc3RyZWFtLWZvcm1hdD0oc3RyaW5nKWJ5dGUtc3RyZWFtLHByb2ZpbGU9KHN0cmluZyltYWluICEgcXVldWUgbWF4LXNpemUtdGltZT0wIG1heC1zaXplLWJ5dGVzPTAgbWF4LXNpemUtYnVmZmVycz0zICEgYXBwc2luayBuYW1lPWFwcHNpbmsi")
	json.Unmarshal(bytes1, &videoPipelineString)
	videopipeline,err := CreatePipeline(videoPipelineString, chans.Confs["adaptive"], chans.Confs["manual"])
	if err != nil {
		fmt.Printf("error initiate video pipeline %s",err.Error())
		return
	}

	i := 0
	videopipeline.RegisterRTPHandler("test",func(pkt *rtp.Packet) {
		if i == 1000 {
			fmt.Printf("packet from %s %s \n","test",pkt.String())
			i = 0
		}

		i++
	})

	y := 0
	videopipeline.RegisterRTPHandler("test2",func(pkt *rtp.Packet) {
		if y == 1000 {
			fmt.Printf("packet from %s %s \n","test2",pkt.String())
			y = 0
		}

		y++
	})

	x := 0
	videopipeline.RegisterRTPHandler("test3",func(pkt *rtp.Packet) {
		if x == 1000 {
			fmt.Printf("packet from %s %s \n","test3",pkt.String())
			x = 0
		}

		x++
	})

	videopipeline.Open()

	videopipeline.DeregisterRTPHandler("test")
	time.Sleep(10 * time.Second)
	videopipeline.Close()
	time.Sleep(1 * time.Second)
}