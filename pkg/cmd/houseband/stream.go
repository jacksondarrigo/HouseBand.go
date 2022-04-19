package houseband

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os/exec"
	"strconv"

	"layeh.com/gopus"
)

const (
	channels  int = 2                   // 1 for mono, 2 for stereo
	frameRate int = 48000               // audio sampling rate
	frameSize int = 960                 // uint16 size of each audio frame
	maxBytes  int = (frameSize * 2) * 2 // max size of opus data
)

type stream struct {
	*gopus.Encoder
	url   string
	audio chan []byte
}

func newStream(url string) *stream {
	opusEncoder, err := gopus.NewEncoder(frameRate, channels, gopus.Audio)
	if err != nil {
		fmt.Println("NewEncoder Error: ", err)
		return nil
	}
	return &stream{opusEncoder, url, make(chan []byte, 2)}
}

func (stream *stream) get() {
	ffmpeg := exec.Command("ffmpeg", "-i", stream.url, "-f", "s16le", "-ar", strconv.Itoa(frameRate), "-ac", strconv.Itoa(channels), "pipe:1")
	ffmpegStdOut, err := ffmpeg.StdoutPipe()
	if err != nil {
		fmt.Println("StdoutPipe Error: ", err)
		return
	}
	ffmpegBuffer := bufio.NewReaderSize(ffmpegStdOut, 16384)
	err = ffmpeg.Start()
	if err != nil {
		fmt.Println("ExecStart Error: ", err)
		return
	}
	for {
		pcmBytes := make([]int16, frameSize*channels)
		err := binary.Read(ffmpegBuffer, binary.LittleEndian, &pcmBytes)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			fmt.Println("EOF")
			break
		}
		if err != nil {
			fmt.Println("error reading from ffmpeg stdout: ", err)
			break
		}
		opusBytes, err := stream.Encode(pcmBytes, frameSize, maxBytes)
		if err != nil {
			fmt.Println("Encoding Error: ", err)
			break
		}
		stream.audio <- opusBytes
	}
	ffmpeg.Process.Kill()
	close(stream.audio)
}
