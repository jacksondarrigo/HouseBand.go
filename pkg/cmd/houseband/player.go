package houseband

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

const (
	channels  int = 2                   // 1 for mono, 2 for stereo
	frameRate int = 48000               // audio sampling rate
	frameSize int = 960                 // uint16 size of each audio frame
	maxBytes  int = (frameSize * 2) * 2 // max size of opus data
)

type musicPlayer struct {
	*discordgo.VoiceConnection
	queue      chan ytRequest
	opusStream *opusStream
	stop       chan bool
}

func newMusicPlayer(voiceChannel *discordgo.VoiceState) *musicPlayer {
	opusStream := newOpusStream()
	return &musicPlayer{&discordgo.VoiceConnection{}, make(chan ytRequest, 24), opusStream, make(chan bool)}
}

//
// Main player loop
//
func (player *musicPlayer) start() {
	err := player.Speaking(true)
	if err != nil {
		fmt.Println("Couldn't set speaking: ", err)
		return
	}
	for {
		nextSong := <-player.queue
		nextSong.nowPlaying()
		player.play(nextSong.Formats.FindByItag(251).URL)
		if len(player.queue) < 1 {
			break
		}
	}
	err = player.Speaking(false)
	if err != nil {
		fmt.Println("Couldn't stop speaking: ", err)
	}
	player.stop <- true
}

func (player *musicPlayer) play(url string) {

	player.opusStream.stream(url)
	for {
		if !player.Ready || player.OpusSend == nil {
			break
		}
		opus, ok := <-player.opusStream.opusAudio
		if !ok {
			break
		}
		player.OpusSend <- opus
	}
}

// func (player *musicPlayer) getAudioStream(url string, pcmAudio chan<- []int16) {
// 	ffmpeg := exec.Command("ffmpeg", "-i", url, "-f", "s16le", "-ar", strconv.Itoa(frameRate), "-ac", strconv.Itoa(channels), "pipe:1")
// 	ffmpegStdOut, err := ffmpeg.StdoutPipe()
// 	if err != nil {
// 		fmt.Println("StdoutPipe Error: ", err)
// 		return
// 	}
// 	ffmpegBuffer := bufio.NewReaderSize(ffmpegStdOut, 16384)
// 	err = ffmpeg.Start()
// 	if err != nil {
// 		fmt.Println("ExecStart Error: ", err)
// 		return
// 	}
// 	for {
// 		audiobuf := make([]int16, frameSize*channels)
// 		err := binary.Read(ffmpegBuffer, binary.LittleEndian, &audiobuf)
// 		if err == io.EOF || err == io.ErrUnexpectedEOF {
// 			//fmt.Println("EOF")
// 			break
// 		}
// 		if err != nil {
// 			fmt.Println("error reading from ffmpeg stdout: ", err)
// 			break
// 		}
// 		pcmAudio <- audiobuf
// 	}
// 	ffmpeg.Process.Kill()
// 	close(pcmAudio)
// }

// func (player *musicPlayer) encodeToOpus(pcmAudio <-chan []int16, opusAudio chan<- []byte) {
// 	opusEncoder, err := gopus.NewEncoder(frameRate, channels, gopus.Audio)
// 	if err != nil {
// 		fmt.Println("NewEncoder Error: ", err)
// 		return
// 	}
// 	for {
// 		pcm, ok := <-pcmAudio
// 		if !ok {
// 			break
// 		}
// 		opus, err := opusEncoder.Encode(pcm, frameSize, maxBytes)
// 		if err != nil {
// 			fmt.Println("Encoding Error: ", err)
// 			break
// 		}
// 		opusAudio <- opus
// 	}
// 	close(opusAudio)
// }
