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
	options = append(options,map[string]string {
		"loopback":"true",
	})
	options = append(options,map[string]string {
		"loopback":"false",
	})

	str := formatDeviceID(source.AudioSource.DeviceID)

	result := false
	var testcase *exec.Cmd
	for _,i := range options{
		testcase = exec.Command("gst-launch-1.0.exe", 
							"wasapisrc", "name=source",fmt.Sprintf("loopback=%s",i["loopback"]),fmt.Sprintf("device=%s",str),
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