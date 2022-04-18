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

type opusStream struct {
	*gopus.Encoder
	pcmAudio  chan []int16
	opusAudio chan []byte
}

func newOpusStream() *opusStream {
	opusEncoder, err := gopus.NewEncoder(frameRate, channels, gopus.Audio)
	if err != nil {
		fmt.Println("NewEncoder Error: ", err)
		return nil
	}
	return &opusStream{opusEncoder, make(chan []int16, 2), make(chan []byte, 2)}
}

func (stream *opusStream) stream(url string) {
	go func() {
		ffmpeg := exec.Command("ffmpeg", "-i", url, "-f", "s16le", "-ar", strconv.Itoa(frameRate), "-ac", strconv.Itoa(channels), "pipe:1")
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
			audiobuf := make([]int16, frameSize*channels)
			err := binary.Read(ffmpegBuffer, binary.LittleEndian, &audiobuf)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				//fmt.Println("EOF")
				break
			}
			if err != nil {
				fmt.Println("error reading from ffmpeg stdout: ", err)
				break
			}
			stream.pcmAudio <- audiobuf
		}
		ffmpeg.Process.Kill()
	}()
	go func() {
		for {
			pcm, ok := <-stream.pcmAudio
			if !ok {
				break
			}
			opus, err := stream.Encode(pcm, frameSize, maxBytes)
			if err != nil {
				fmt.Println("Encoding Error: ", err)
				break
			}
			stream.opusAudio <- opus
		}
	}()
}

func (stream *opusStream) get(url string) {
	ffmpeg := exec.Command("ffmpeg", "-i", url, "-f", "s16le", "-ar", strconv.Itoa(frameRate), "-ac", strconv.Itoa(channels), "pipe:1")
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
		audiobuf := make([]int16, frameSize*channels)
		err := binary.Read(ffmpegBuffer, binary.LittleEndian, &audiobuf)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			//fmt.Println("EOF")
			break
		}
		if err != nil {
			fmt.Println("error reading from ffmpeg stdout: ", err)
			break
		}
		stream.pcmAudio <- audiobuf
	}
	ffmpeg.Process.Kill()
	//close(pcmAudio)
}

func (stream *opusStream) encode() {
	for {
		pcm, ok := <-stream.pcmAudio
		if !ok {
			break
		}
		opus, err := stream.Encode(pcm, frameSize, maxBytes)
		if err != nil {
			fmt.Println("Encoding Error: ", err)
			break
		}
		stream.opusAudio <- opus
	}
	//close(opusAudio)
}
