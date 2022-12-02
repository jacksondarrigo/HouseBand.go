package player

import (
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/jacksondarrigo/HouseBand.go/cmd/houseband/request"
	"github.com/jacksondarrigo/HouseBand.go/cmd/houseband/stream"
)

type MusicPlayer struct {
	queue
	*discordgo.VoiceConnection
	Messages    chan Message
	CurrentSong *request.Request
	Started     bool
	Stop        chan bool
	Next        chan bool
}

type queue struct {
	sync.Mutex
	Queue []*request.Request
}

type Message struct {
	ChannelId string
	Content   string
}

func New() *MusicPlayer {
	return &MusicPlayer{queue{Queue: make([]*request.Request, 0, 100)}, &discordgo.VoiceConnection{}, make(chan Message), nil, true, make(chan bool), make(chan bool)}
}

func (player *MusicPlayer) AddToQueue(request *request.Request) {
	player.queue.Lock()
	defer player.queue.Unlock()
	player.Queue = append(player.Queue, request)
	player.Messages <- Message{ChannelId: request.InteractionChannel, Content: "*Added to Queue:* `" + request.Title + "`"}
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
		streamUrl, err := player.CurrentSong.GetStream()
		if err != nil {
			player.Messages <- Message{ChannelId: player.CurrentSong.InteractionChannel, Content: "**Error Playing:** `" + player.CurrentSong.Title + "`; *skipping song*"}
			continue
		}
		player.Messages <- Message{ChannelId: player.CurrentSong.InteractionChannel, Content: "**Now Playing:** `" + player.CurrentSong.Title + "`"}
		player.play(streamUrl)
	}
	err = player.Speaking(false)
	if err != nil {
		log.Println("Error: Couldn't stop speaking: ", err)
	}
}

func (player *MusicPlayer) play(streamUrl string) {
	stream := stream.New(streamUrl)
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
	player.queue.Lock()
	defer player.queue.Unlock()
	request := player.Queue[0]
	player.Queue = player.Queue[1:]
	return request
}

func (player *MusicPlayer) isEmpty() bool {
	return len(player.Queue) < 1
}
