package gsttest

import (
	"fmt"
	"os/exec"
	"time"

	childprocess "github.com/OnePlay-Internet/webrtc-proxy/util/child-process"
	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
)


func formatDeviceID(in string) string {

	modified := make([]byte,0);
	byts := []byte(in);

	for index,byt := range byts {
		if byts[index] == []byte("{")[0] || 
		   byts[index] == []byte("}")[0] {
			modified = append(modified, []byte("\\")[0]);	
		}
		modified = append(modified, byt);	
	}

	ret := []byte("\"")
	ret = append(ret , modified...)
	ret = append(ret ,[]byte("\"")...)
	return string(ret)
}


func GstTestAudio(source *config.ListenerConfig) string{
	options := make([]map[string]string,0); 

	for _,soundcard := range source.AudioSource {
		if soundcard.Api == "wasapi" {
			options = append(options,map[string]string { 
				"element":"wasapisrc", 
				"loopback": "true",
				"device": formatDeviceID(soundcard.DeviceID),
			})
		} else if soundcard.Api == "wasapi2" && soundcard.IsDefault {
			var loopback string
			if soundcard.IsLoopback { 
				loopback = "true" 
			} else { 
				loopback = "false" 
			}

			options = append(options,map[string]string { 
				"element":"wasapi2src", 
				"loopback": loopback,
				"device": formatDeviceID(soundcard.DeviceID),
			})
		}
	}


	result := false
	var testcase *exec.Cmd
	for _,i := range options{
		testcase = exec.Command("gst-launch-1.0.exe", 
							i["element"], "name=source",fmt.Sprintf("loopback=%s",i["loopback"]),fmt.Sprintf("device=%s",i["device"]),
							"!","queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3","!",
							"audioconvert",
							"!","queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3","!",
							"opusenc",fmt.Sprintf("bitrate=%d",source.Bitrate),
							"!","queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3","!",
							"appsink","name=appsink")

		done := make(chan bool)
		failed := make(chan bool)
		success := make(chan bool)
		go func ()  {
			childprocess.HandleProcess(testcase);
			failed<-true;
		}()
		go func ()  {
			time.Sleep(3 * time.Second);
			success<-true;
		}()
		go func ()  {
			for {
				select {
				case <-success:
					result = true;
					done<-true;
					return;
				case <-failed:
					result = false;
					done<-true;
					return;
				}
			}
		}()
		<-done
		if testcase.Process != nil {
			testcase.Process.Kill()
		}

		if result {
			break;
		}
	}


	log := make([]byte,0);
	for _,i := range testcase.Args[1:] {
		log = append(log, append([]byte(i),[]byte(" ")...)...);
	}

	if result {
		return string(log)
	} else {
		return "";
	}

}