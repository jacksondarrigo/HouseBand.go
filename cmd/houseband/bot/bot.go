package bot

import (
	"encoding/json"
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

	bot.startWebServer()
	defer bot.stopWebServer()

	fmt.Println("HouseBandTest is now running.  Press CTRL-C to exit.")
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-signals

	// Remove all existing commands on exit - prevents users from issuing a command when the bot is unavailable, also make sures old commands are deleted when removed from the codebase
	//bot.deleteCommands()
}

func (bot *Bot) stopWebServer() {
	err := bot.webServer.Close()
	if err != nil {
		fmt.Println("Error closing web server: ", err)
	}
}

func (bot *Bot) startWebServer() {
	mux := http.NewServeMux()
	webHandler := func(w http.ResponseWriter, r *http.Request) {
		body, err := json.Marshal(bot.Session.State)
		if err != nil {
			body = []byte("Error retrieving bot state: " + err.Error())
		}
		fmt.Fprintf(w, string(body))
		// var guilds []string = []string{}
		// for _, guild := range bot.State.Guilds {
		// 	guilds = append(guilds, guild.Name)
		// }
		// body, err := json.Marshal(guilds)
		// if err != nil {
		// 	body = []byte(strings.Join(guilds, "\n "))
		// }
		// fmt.Fprintf(w, "Guilds:\n"+string(body)+"\n")
		// var voiceConnections []string = []string{}
		// for _, vc := range bot.VoiceConnections {
		// 	guild, err := bot.Guild(vc.GuildID)
		// 	if err != nil {
		// 		continue
		// 	}
		// 	channel, err := bot.Channel(vc.ChannelID)
		// 	if err != nil {
		// 		continue
		// 	}
		// 	voiceConnections = append(voiceConnections, guild.Name+": "+channel.Name)
		// }
		// body, err = json.Marshal(voiceConnections)
		// if err != nil {
		// 	body = []byte(strings.Join(voiceConnections, "\n "))
		// }
		// fmt.Fprintf(w, "\nVoice Connections:\n"+string(body)+"\n")
	}
	mux.HandleFunc("/", webHandler)
	bot.webServer = &http.Server{Addr: ":8080", Handler: mux}
	go func() {
		err := bot.webServer.ListenAndServe()
		if err != nil {
			fmt.Println("Error creating web server: ", err)
		}
	}()
}
