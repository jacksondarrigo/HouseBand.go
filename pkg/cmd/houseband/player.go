package houseband

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os/exec"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"layeh.com/gopus"
)

const (
	channels  int = 2                   // 1 for mono, 2 for stereo
	frameRate int = 48000               // audio sampling rate
	frameSize int = 960                 // uint16 size of each audio frame
	maxBytes  int = (frameSize * 2) * 2 // max size of opus data
)

type musicPlayer struct {
	*discordgo.VoiceConnection
	queue      chan songRequest
	nowPlaying chan songRequest
	stop       chan bool
}

func (bot *Bot) newMusicPlayer(voiceChannel *discordgo.VoiceState) *musicPlayer {
	//var queue Queue = Queue{}
	var voiceConn *discordgo.VoiceConnection
	player := &musicPlayer{voiceConn, make(chan songRequest, 24), make(chan songRequest), make(chan bool)}
	go func() {
		var err error
		voiceConn, err = bot.ChannelVoiceJoin(voiceChannel.GuildID, voiceChannel.ChannelID, false, false)
		if err != nil {
			fmt.Println("Error while joining channel: ", err)
		}
		player.VoiceConnection = voiceConn
		player.startPlayer()
	}()
	return player
}

//
// Main player loop
//
func (player *musicPlayer) startPlayer() {
	for {
		nextSong, ok := <-player.queue
		if !ok {
			break
		}
		fmt.Println("received ", nextSong, " on ", player)
		player.nowPlaying <- nextSong
		player.playAudio(nextSong.playURL)
	}
	player.stop <- true
	player.Disconnect()
}

func (player *musicPlayer) playAudio(url string) {

	fmt.Println("playing")
	pcmAudio := make(chan []int16, 2)
	opusAudio := make(chan []byte, 2)
	go func() {
		fmt.Println("getting")
		player.getAudio(url, pcmAudio)
	}()
	go func() {
		fmt.Println("encoding")
		player.encodeAudio(pcmAudio, opusAudio)
	}()
	fmt.Println("sending")
	player.sendAudio(opusAudio)
}

func (player *musicPlayer) encodeAudio(input <-chan []int16, output chan<- []byte) {
	opusEncoder, err := gopus.NewEncoder(frameRate, channels, gopus.Audio)
	if err != nil {
		fmt.Println("NewEncoder Error: ", err)
		return
	}
	for {
		pcm, ok := <-input
		if !ok {
			break
		}
		opus, err := opusEncoder.Encode(pcm, frameSize, maxBytes)
		if err != nil {
			fmt.Println("Encoding Error: ", err)
			break
		}
		output <- opus
	}
	close(output)
}

func (player *musicPlayer) sendAudio(opusAudio <-chan []byte) {
	// i := 0
	// for {
	// 	if !player.Ready {
	// 		time.Sleep(1 * time.Second)
	// 		i++
	// 	} else {
	// 		break
	// 	}
	// 	if i > 10 {
	// 		return
	// 	}
	// }
	fmt.Println("player is ", player.Ready)
	fmt.Println("setting speaking on ", player.ChannelID)
	err := player.Speaking(true)
	if err != nil {
		fmt.Println("Couldn't set speaking: ", err)
	}
	for {
		// if !player.Ready || player.OpusSend == nil {
		// 	continue
		// }

		opus, ok := <-opusAudio
		if !ok {
			break
		}
		// TODO: needed?
		// if !player.Ready || player.OpusSend == nil {
		// 	break
		// }
		player.OpusSend <- opus
	}
	//player.stop <- true
	err = player.Speaking(false)
	if err != nil {
		fmt.Println("Couldn't stop speaking: ", err)
	}
}

func (player *musicPlayer) getAudio(url string, pcmAudio chan<- []int16) {
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
		} else {
			if err != nil {
				fmt.Println("error reading from ffmpeg stdout: ", err)
				break
			}
		}
		pcmAudio <- audiobuf
	}
	ffmpeg.Process.Kill()
	close(pcmAudio)
}
