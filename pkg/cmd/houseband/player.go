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
	queue        Queue
	voiceChannel *discordgo.VoiceState
	voiceConn    *discordgo.VoiceConnection
	bot          *Bot
	stop         chan bool
	started      bool
}

func newMusicPlayer(voiceChannel *discordgo.VoiceState) *musicPlayer {
	var queue Queue = Queue{make(chan *songRequest, 24)}
	var voiceCon *discordgo.VoiceConnection
	var bot *Bot
	return &musicPlayer{queue, voiceChannel, voiceCon, bot, make(chan bool), false}
}

func (player *musicPlayer) startPlayer() {
	player.started = true
	var err error
	player.voiceConn, err = player.bot.ChannelVoiceJoin(player.voiceChannel.GuildID, player.voiceChannel.ChannelID, false, false)
	if err != nil {
		fmt.Println("Error while joining channel: ", err)
	}
	for {
		if player.queue.length() < 1 {
			break
		}
		nextSong := player.queue.dequeue()
		player.bot.ChannelMessageSend(nextSong.requestChannelID, "**Now Playing:** `"+nextSong.title+"`")
		player.playAudio(nextSong.playURL)
	}
	delete(player.bot.musicPlayers, player.voiceConn.GuildID)
	player.voiceConn.Disconnect()
}

func (player *musicPlayer) playAudio(url string) {
	pcmAudio := make(chan []int16, 2)
	opusAudio := make(chan []byte, 2)
	go func() {
		player.getAudio(url, pcmAudio)
	}()
	go func() {
		player.encodeAudio(pcmAudio, opusAudio)
	}()
	go func() {
		player.sendAudio(opusAudio)
	}()
	<-player.stop
}

func (player *musicPlayer) encodeAudio(pcmAudio <-chan []int16, opusAudio chan<- []byte) {
	opusEncoder, err := gopus.NewEncoder(frameRate, channels, gopus.Audio)
	if err != nil {
		fmt.Println("NewEncoder Error: ", err)
		return
	}
	for {
		pcm, ok := <-pcmAudio
		if !ok {
			break
		}
		opus, err := opusEncoder.Encode(pcm, frameSize, maxBytes)
		if err != nil {
			fmt.Println("Encoding Error: ", err)
			break
		}
		opusAudio <- opus
	}
	close(opusAudio)
}

func (player *musicPlayer) sendAudio(opusAudio <-chan []byte) {
	err := player.voiceConn.Speaking(true)
	if err != nil {
		fmt.Println("Couldn't set speaking: ", err)
	}
	for {
		opus, ok := <-opusAudio
		if !ok {
			break
		}
		// TODO: needed?
		if !player.voiceConn.Ready || player.voiceConn.OpusSend == nil {
			break
		}
		player.voiceConn.OpusSend <- opus
	}
	player.stop <- true
	err = player.voiceConn.Speaking(false)
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
