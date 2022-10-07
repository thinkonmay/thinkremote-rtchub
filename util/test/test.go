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
		   byts[index] == []byte("?")[0] ||
		   byts[index] == []byte("#")[0] ||
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
	soundcard := source.AudioSource

	// wasapi2 has higher priority
	if soundcard.Api == "wasapi2" {
		options = append(options,map[string]string { 
			"element":"wasapi2src", 
			"device": formatDeviceID(soundcard.DeviceID),
		})
	} else if soundcard.Api == "wasapi" {
		options = append(options,map[string]string { 
			"element":"wasapisrc", 
			"device": formatDeviceID(soundcard.DeviceID),
		})
	} 

	if len(options) == 0 {
		return ""
	}


	result := false
	var testcase *exec.Cmd
	for _,i := range options{
		testcase = exec.Command("gst-launch-1.0.exe", 
							i["element"], "name=source","loopback=true",fmt.Sprintf("device=%s",i["device"]),
							"!","queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3","!",
							"audioconvert",
							"!","queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3","!",
							"opusenc",fmt.Sprintf("bitrate=%d",source.Bitrate), "name=encoder",
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
			time.Sleep(2 * time.Second);
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



	if result {
		log := make([]byte,0);
		for _,i := range testcase.Args[1:] {
			log = append(log, append([]byte(i),[]byte(" ")...)...);
		}
		return string(log)
	} else {
		return "";
	}
}


func GstTestNvCodec(source *config.ListenerConfig) string{
	testcase := exec.Command("gst-launch-1.0.exe", "d3d11screencapturesrc","blocksize=8192",
						fmt.Sprintf("monitor-handle=%d",source.VideoSource.MonitorHandle),
						"!", fmt.Sprintf("video/x-raw(memory:D3D11Memory),framerate=%d/1",source.VideoSource.Framerate), "!",
						"!","queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3","!",
						"d3d11download",
						"!","queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3","!",
						"nvh264enc",fmt.Sprintf("bitrate=%d",source.Bitrate),"zerolatency=true","rc-mode=2","name=encoder",
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
		time.Sleep(2 * time.Second);
		success<-true;
	}()

	var result bool;
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
		log := make([]byte,0);
		for _,i := range testcase.Args[1:] {
			log = append(log, append([]byte(i),[]byte(" ")...)...);
		}
		return string(log)
	} else {
		return "";
	}
}



func GstTestMediaFoundation(source *config.ListenerConfig) string{
	testcase := exec.Command("gst-launch-1.0.exe", "d3d11screencapturesrc","blocksize=8192",
						fmt.Sprintf("monitor-handle=%d",source.VideoSource.MonitorHandle),
						"!", fmt.Sprintf("video/x-raw(memory:D3D11Memory),framerate=%d/1",source.VideoSource.Framerate), "!",
						"!","queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3","!",
						"d3d11convert",
						"!","queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3","!",
						"mfh264enc",fmt.Sprintf("bitrate=%d",source.Bitrate),"rc-mode=0","low-latency=true","ref=1","quality-vs-speed=0","name=encoder",
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
		time.Sleep(2 * time.Second);
		success<-true;
	}()

	var result bool;
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
		log := make([]byte,0);
		for _,i := range testcase.Args[1:] {
			log = append(log, append([]byte(i),[]byte(" ")...)...);
		}
		return string(log)
	} else {
		return "";
	}
}


func GstTestSoftwareEncoder(source *config.ListenerConfig) string{
	testcase := exec.Command("gst-launch-1.0.exe", "d3d11screencapturesrc","blocksize=8192",
						fmt.Sprintf("monitor-handle=%d",source.VideoSource.MonitorHandle),
						"!", fmt.Sprintf("video/x-raw,framerate=%d/1",source.VideoSource.Framerate), "!",
						"!","queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3","!",
						"d3d11convert",
						"!","queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3","!",
						"d3d11download",
						"!","queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3","!",
						"openh264enc",fmt.Sprintf("bitrate=%d",source.Bitrate),"usage-type=1","rate-control=1","multi-thread=8","name=encoder",
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
		time.Sleep(2 * time.Second);
		success<-true;
	}()

	var result bool;
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
		log := make([]byte,0);
		for _,i := range testcase.Args[1:] {
			log = append(log, append([]byte(i),[]byte(" ")...)...);
		}
		return string(log)
	} else {
		return "";
	}
}