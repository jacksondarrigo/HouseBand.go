package houseband

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type musicPlayer struct {
	*discordgo.VoiceConnection
	queue   chan request
	stop    chan bool
	next    chan bool
	started bool
}

func newMusicPlayer() *musicPlayer {
	return &musicPlayer{&discordgo.VoiceConnection{}, make(chan request, 24), make(chan bool), make(chan bool), false}
}

//
// Main player loop
//
func (player *musicPlayer) run() {
	player.started = true
	err := player.Speaking(true)
	if err != nil {
		fmt.Println("Couldn't set speaking: ", err)
		return
	}
	for {
		nextSong := <-player.queue
		nextSong.nowPlaying()
		stream := newStream(nextSong.Formats.FindByItag(251).URL)
		player.play(stream)
		if player.isQueueEmpty() || !player.started {
			break
		}
	}
	err = player.Speaking(false)
	if err != nil {
		fmt.Println("Couldn't stop speaking: ", err)
	}
	player.Disconnect()
}

func (player *musicPlayer) play(stream *stream) {

	go stream.get()
	for {
		select {
		case opusBytes, ok := <-stream.audio:
			if !ok {
				return
			}
			player.OpusSend <- opusBytes
		case <-player.next:
			return
		case <-player.stop:
			player.started = false
			return
		}
	}
}

func (player *musicPlayer) isQueueEmpty() bool {
	return len(player.queue) < 1
}
