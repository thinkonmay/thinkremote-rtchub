package gsttest

import (
	"fmt"
	"os/exec"
	"time"

	childprocess "github.com/OnePlay-Internet/webrtc-proxy/util/child-process"
	"github.com/OnePlay-Internet/webrtc-proxy/util/tool"
)

const (
	VideoClockRate = 90000
	AudioClockRate = 48000

	defaultAudioBitrate = 128000
	defaultVideoBitrate = 6000
)









func FindTestCmd(plugin string, handle int, DeviceID string) *exec.Cmd{
	switch plugin {
	case "media foundation":
	return exec.Command("gst-launch-1.0.exe", "d3d11screencapturesrc", "blocksize=8192", "do-timestamp=true",
		fmt.Sprintf("monitor-handle=%d", handle),
		"!", "capsfilter", "name=framerateFilter",
		"!", fmt.Sprintf("video/x-raw(memory:D3D11Memory),clock-rate=%d", VideoClockRate),
		"!", "queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3", "!",
		"d3d11convert",
		"!", "queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3", "!",
		"mfh264enc", fmt.Sprintf("bitrate=%d", defaultVideoBitrate), "gop-size=6", "rc-mode=0", "low-latency=true", "ref=1", "quality-vs-speed=0", "name=encoder",
		"!", "queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3", "!",
		"appsink", "name=appsink")
	case "nvcodec":
	return exec.Command("gst-launch-1.0.exe", "d3d11screencapturesrc", "blocksize=8192", "do-timestamp=true",
		fmt.Sprintf("monitor-handle=%d", handle),
		"!", "capsfilter", "name=framerateFilter",
		"!", fmt.Sprintf("video/x-raw(memory:D3D11Memory),clock-rate=%d", VideoClockRate),
		"!", "queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3", "!",
		"cudaupload",
		"!", "queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3", "!",
		"nvh264enc", fmt.Sprintf("bitrate=%d", defaultVideoBitrate), "zerolatency=true", "rc-mode=2", "name=encoder",
		"!", "queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3", "!",
		"h264parse", "config-interval=-1",
		"!", "queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3", "!",
		"appsink", "name=appsink")
	case "quicksync":
	return exec.Command("gst-launch-1.0.exe", "d3d11screencapturesrc", "blocksize=8192", "do-timestamp=true",
		fmt.Sprintf("monitor-handle=%d", handle),
		"!", "capsfilter", "name=framerateFilter",
		"!", fmt.Sprintf("video/x-raw(memory:D3D11Memory),clock-rate=%d", VideoClockRate),
		"!", "queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3", "!",
		"d3d11convert",
		"!", "queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3", "!",		
		"qsvh264enc", fmt.Sprintf("bitrate=%d", defaultVideoBitrate), "rate-control=1", "gop-size=6","ref-frames=1" ,"low-latency=true","target-usage=7" ,"name=encoder",
		"!", "queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3", "!",
		"appsink", "name=appsink")
	case "amf":
	return exec.Command("gst-launch-1.0.exe", "d3d11screencapturesrc", "blocksize=8192", "do-timestamp=true",
		fmt.Sprintf("monitor-handle=%d", handle),
		"!", "capsfilter", "name=framerateFilter",
		"!", fmt.Sprintf("video/x-raw(memory:D3D11Memory),clock-rate=%d", VideoClockRate),
		"!", "queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3", "!",
		"d3d11convert",
		"!", "queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3", "!",		
		"amfh264enc", fmt.Sprintf("bitrate=%d", defaultVideoBitrate), "rate-control=1", "gop-size=6","usage=1","name=encoder",
		"!", "queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3", "!",
		"h264parse", "config-interval=-1",
		"!", "queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3", "!",
		"appsink", "name=appsink")
	case "opencodec":
	return exec.Command("gst-launch-1.0.exe", "d3d11screencapturesrc", "blocksize=8192", "do-timestamp=true",
		fmt.Sprintf("monitor-handle=%d", handle),
		"!", "capsfilter", "name=framerateFilter",
		"!", fmt.Sprintf("video/x-raw(memory:D3D11Memory),clock-rate=%d", VideoClockRate),
		"!", "queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3", "!",
		"d3d11convert",
		"!", "queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3", "!",
		"d3d11download",
		"!", "queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3", "!",
		"openh264enc", fmt.Sprintf("bitrate=%d", defaultVideoBitrate), "usage-type=1", "rate-control=1", "multi-thread=8", "name=encoder",
		"!", "queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3", "!",
		"appsink", "name=appsink")
	case "wasapi2":
	return exec.Command("gst-launch-1.0.exe", "wasapi2src", "name=source", "loopback=true", 
		fmt.Sprintf("device=%s", formatAudioDeviceID(DeviceID)),
		"!", "audio/x-raw",
		"!", "queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3", "!",
		"audioresample",
		"!", fmt.Sprintf("audio/x-raw,clock-rate=%d", AudioClockRate),
		"!", "queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3", "!",
		"audioconvert",
		"!", "queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3", "!",
		"opusenc", fmt.Sprintf("bitrate=%d", defaultAudioBitrate), "name=encoder",
		"!", "queue", "max-size-time=0", "max-size-bytes=0", "max-size-buffers=3", "!",
		"appsink", "name=appsink")
	default:
		return nil
	}
}



func formatAudioDeviceID(in string) string {

	modified := make([]byte, 0)
	byts := []byte(in)

	for index, byt := range byts {
		if byts[index] == []byte("{")[0] ||
			byts[index] == []byte("?")[0] ||
			byts[index] == []byte("#")[0] ||
			byts[index] == []byte("}")[0] {
			modified = append(modified, []byte("\\")[0])
		}
		modified = append(modified, byt)
	}

	ret := []byte("\"")
	ret = append(ret, modified...)
	ret = append(ret, []byte("\"")...)
	return string(ret)
}
func filterWithClass(available map[string]string, classes ...[]string) string {
	for _,class := range classes {
		for _,candidate := range class {
			if available[candidate] != "" {
				return available[candidate];
			}
		}
	}

	return ""
}

func GstTestAudio(video *tool.Soundcard) string {
	testcase := FindTestCmd(video.Api,0,video.DeviceID)
	return gstTestGeneric(video.Api,testcase)
}



func GstTestVideo(video *tool.Monitor) string {

	// TODO
	video_plugins := []string{"nvcodec","amf","quicksync", "media foundation","opencodec"};
	// video_plugins := []string{"quicksync","media foundation","opencodec"};

	class1 := []string{"amf","nvcodec","quicksync" };
	class2 := []string{"media foundation" };
	class3 := []string{"opencodec" };

	available_pipelines := make(map[string]string)

	testAll := false
	count := 0 
	for _,_plugin := range video_plugins {
		testcase := FindTestCmd(_plugin,video.MonitorHandle,"")
		pipeline := gstTestGeneric(_plugin,testcase)
		if pipeline != "" {
			available_pipelines[_plugin] = pipeline;
			if !testAll {
				return pipeline
			}
		} 
		count++
	}

	for { if count == len(video_plugins) { break; }
		time.Sleep(100 * time.Millisecond)
	}

	return filterWithClass(available_pipelines,class1,class2,class3);
}


func gstTestGeneric(plugin string,testcase *exec.Cmd) string {
	done, failed, success := make(chan bool),make(chan bool),make(chan bool)

	go func() {
		childprocess.HandleProcess(testcase)
		failed <- true
	}()
	go func() {
		time.Sleep(3 * time.Second)
		success <- true
	}()

	var result bool
	go func() {
		for {
			select {
			case <-success:
				result = true
				done <- true
				return
			case <-failed:
				result = false
				done <- true
				return
			}
		}
	}()
	<-done
	if testcase.Process != nil {
		testcase.Process.Kill()
	}

	if result {
		log := make([]byte, 0)
		for _, i := range testcase.Args[1:] {
			log = append(log, append([]byte(i), []byte(" ")...)...)
		}
		return string(log)
	} else {
		return ""
	}
}
