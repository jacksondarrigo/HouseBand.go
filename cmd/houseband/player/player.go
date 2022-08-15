package player

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jacksondarrigo/HouseBand.go/cmd/houseband/request"
)

type MusicPlayer struct {
	*discordgo.VoiceConnection
	Queue       []*request.Request
	CurrentSong *request.Request
	Stop        chan bool
	Next        chan bool
	Started     bool
}

func NewMusicPlayer() *MusicPlayer {
	return &MusicPlayer{&discordgo.VoiceConnection{}, make([]*request.Request, 0, 24), nil, make(chan bool), make(chan bool), false}
}

// Main player loop
func (player *MusicPlayer) Run() {
	player.Started = true
	err := player.Speaking(true)
	if err != nil {
		fmt.Println("Couldn't set speaking: ", err)
		return
	}
	for player.Started && !player.QueueEmpty() {
		player.CurrentSong = player.NextSong()
		player.CurrentSong.NowPlaying()
		player.Play(player.CurrentSong.StreamURL)
	}
	err = player.Speaking(false)
	if err != nil {
		fmt.Println("Couldn't stop speaking: ", err)
	}
	player.Disconnect()
}

func (player *MusicPlayer) Play(url string) {
	stream := newStream(url)
	go stream.get()
	for {
		select {
		case opusBytes, ok := <-stream.audio:
			if !ok {
				return
			}
			player.OpusSend <- opusBytes
		case <-player.Next:
			return
		case <-player.Stop:
			player.Started = false
			return
		}
	}
}

func (player *MusicPlayer) AddToQueue(request *request.Request) {
	player.Queue = append(player.Queue, request)
}

func (player *MusicPlayer) NextSong() *request.Request {
	request := player.Queue[0]
	player.Queue = player.Queue[1:]
	return request
}

func (player *MusicPlayer) QueueEmpty() bool {
	return len(player.Queue) < 1
}
