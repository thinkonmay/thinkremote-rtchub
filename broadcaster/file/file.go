package file

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"

	"github.com/OnePlay-Internet/webrtc-proxy/broadcaster"
	"github.com/OnePlay-Internet/webrtc-proxy/util/config"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/ivfwriter"
)

type FileSaver struct {
	config  *config.BroadcasterConfig
	writter media.Writer

	ffmpegIn  io.Writer
	ffmpegErr io.Reader
	ffmpegOut io.Reader

	buffer     []byte
	bufferSize int

	closeChannel chan bool
}

const (
	frameX      = 960
	frameY      = 720
	frameSize   = frameX * frameY * 3
	minimumArea = 3000
)

func NewUDPBroadcaster(config *config.BroadcasterConfig) (udp *FileSaver, err error) {
	udp = &FileSaver{}
	ffmpeg := exec.Command("ffmpeg.exe", "-i", "pipe:0", "copy", "-f", "segment", "-segment_atclocktime", "1", ",-segment_time", "900", "-reset_timestamps", "1", "-strftime", "1", "out.mkv")
	udp.ffmpegIn, _ = ffmpeg.StdinPipe()
	udp.ffmpegErr, _ = ffmpeg.StderrPipe()
	udp.ffmpegOut, _ = ffmpeg.StdoutPipe()

	if err := ffmpeg.Start(); err != nil {
		panic(err)
	}
	go func() {
		scanner := bufio.NewScanner(udp.ffmpegErr)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	udp = &FileSaver{}
	udp.config = config
	udp.bufferSize = 10000

	udp.writter, err = ivfwriter.NewWith(udp.ffmpegIn)
	if err != nil {
		return
	}

	udp.buffer = make([]byte, udp.bufferSize)
	udp.closeChannel = make(chan bool)
	return
}

func (udp *FileSaver) Write(packet *rtp.Packet) {
	fmt.Printf("SAVED %s\n", packet.String())

	err := udp.writter.WriteRTP(packet)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
	}
}

func (udp *FileSaver) Close() {
	udp.closeChannel <- true
}

func (udp *FileSaver) OnClose(fun broadcaster.OnCloseFunc) {
	go func() {
		<-udp.closeChannel
		fun(udp)
	}()
}

func (udp *FileSaver) ReadConfig() *config.BroadcasterConfig {
	return udp.config
}
