package bot

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/jacksondarrigo/HouseBand.go/cmd/houseband/player"
)

type Bot struct {
	*discordgo.Session
	mu           sync.Mutex
	musicPlayers map[string]*player.MusicPlayer
	webServer    *http.Server
}

func New(token string) *Bot {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return nil
	}
	return &Bot{session, sync.Mutex{}, make(map[string]*player.MusicPlayer), nil}
}

func (bot *Bot) Run(resetCommands bool) {
	if resetCommands {
		bot.AddHandler(bot.onReadyReset)
	} else {
		bot.AddHandler(bot.onReady)
	}
	bot.AddHandler(bot.interactionHandler)
	bot.Identify.Intents = discordgo.IntentsAllWithoutPrivileged
	err := bot.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
		return
	}
	defer bot.Close()

	fmt.Println("HouseBandTest is now running.  Press CTRL-C to exit.")
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-signals
}
