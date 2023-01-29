package player

import (
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/jacksondarrigo/HouseBand.go/cmd/houseband/request"
	"github.com/jacksondarrigo/HouseBand.go/cmd/houseband/stream"
)

type MusicPlayer struct {
	*discordgo.VoiceConnection
	Queue       []*request.Request
	CurrentSong *request.Request
	Messages    chan Message
	Started     bool
	Stop        chan bool
	Next        chan bool
	Mutex       sync.Mutex
}

type Message struct {
	ChannelId string
	Content   string
}

func New(vc *discordgo.VoiceConnection) *MusicPlayer {
	return &MusicPlayer{vc, make([]*request.Request, 0, 100), nil, make(chan Message), false, make(chan bool), make(chan bool), sync.Mutex{}}
}

func (player *MusicPlayer) AddToQueue(request *request.Request) {
	player.Mutex.Lock()
	defer player.Mutex.Unlock()
	player.Queue = append(player.Queue, request)
	player.Messages <- Message{ChannelId: request.InteractionChannel, Content: "*Added to Queue:* `" + request.Title + "`"}
	if !player.Started {
		player.Started = true
		go func() {
			player.Run()
			player.Disconnect()
			close(player.Messages)
		}()
	}
}

// Main player loop
func (player *MusicPlayer) Run() {
	err := player.VoiceConnection.Speaking(true)
	if err != nil {
		log.Println("Error: Couldn't set speaking: ", err)
		return
	}
	for player.Started && !player.isEmpty() {
		player.CurrentSong = player.nextSong()
		streamUrl, err := player.CurrentSong.GetStreamURL()
		if err != nil {
			player.Messages <- Message{ChannelId: player.CurrentSong.InteractionChannel, Content: "Error: Could not retrieve stream for `" + player.CurrentSong.Title + "`: " + err.Error()}
			continue
		}
		player.Messages <- Message{ChannelId: player.CurrentSong.InteractionChannel, Content: "**Now Playing:** `" + player.CurrentSong.Title + "`"}
		err = player.play(streamUrl)
		if err != nil {
			player.Messages <- Message{ChannelId: player.CurrentSong.InteractionChannel, Content: "Error: There was a problem during playback: " + err.Error()}
		}
	}
	err = player.Speaking(false)
	if err != nil {
		log.Println("Error: Couldn't stop speaking: ", err)
	}
}

func (player *MusicPlayer) play(streamUrl string) (err error) {
	stream, err := stream.New(streamUrl)
	if err != nil {
		return
	}
	go stream.Get()
	for {
		select {
		case opusBytes, ok := <-stream.Audio:
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

func (player *MusicPlayer) nextSong() *request.Request {
	player.Mutex.Lock()
	defer player.Mutex.Unlock()
	request := player.Queue[0]
	player.Queue = player.Queue[1:]
	return request
}

func (player *MusicPlayer) isEmpty() bool {
	return len(player.Queue) < 1
}
