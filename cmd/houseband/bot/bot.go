package bot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/jacksondarrigo/HouseBand.go/cmd/houseband/player"
)

type Bot struct {
	*discordgo.Session
	musicPlayers map[string]*player.MusicPlayer
	webServer    *http.Server
}

func New(token string) *Bot {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return nil
	}
	return &Bot{session, make(map[string]*player.MusicPlayer), nil}
}

func (bot *Bot) Run() {
	bot.AddHandler(bot.onReady)
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

func (bot *Bot) onReady(session *discordgo.Session, event *discordgo.Ready) {
	bot.createCommands()
}

func (bot *Bot) createCommands() {
	commands := []*discordgo.ApplicationCommand{
		{ //  Oh.. oh, song? You want to sing a song? You were excited about singing a song? GOOOOOOOOOOD!
			Name:        "play",
			Description: "Play a song from YouTube",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "url",
					Description: "URL to song",
					Required:    true,
				},
			},
		},
		{ //  What are you doing? Don't do that... I COMMAND YOU TO STOP.
			Name:        "stop",
			Description: "Stop playing music and disconnect",
		},
		{ //  ... I didn't tell you what to- You're skipping a line, dude.
			Name:        "skip",
			Description: "Skip the current song in queue",
		},
		{
			Name:        "queue",
			Description: "List all songs in queue",
		},
	}
	for _, command := range commands {
		_, err := bot.ApplicationCommandCreate(bot.State.User.ID, "", command)
		if err != nil {
			fmt.Println("Error while registering commands: ", err)
		}
	}
}

func (bot *Bot) deleteCommands() {
	commands, err := bot.ApplicationCommands(bot.State.User.ID, "")
	if err != nil {
		fmt.Println("Error while getting commands: ", err)
		return
	}
	for _, command := range commands {
		err = bot.ApplicationCommandDelete(bot.State.User.ID, "", command.ID)
		if err != nil {
			fmt.Println("Error while clearing commands: ", err)
		}
	}
}

func (bot *Bot) interactionHandler(session *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		bot.commandHandler(i)
	case discordgo.InteractionMessageComponent:
		bot.componentHandler(i)
	}
}

// TODO: Implement component handler
func (bot *Bot) componentHandler(i *discordgo.InteractionCreate) {
	return
}

func (bot *Bot) commandHandler(interact *discordgo.InteractionCreate) {
	err := bot.InteractionRespond(interact.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		fmt.Println("Error while sending interaction response: ", err)
		return
	}
	var message string
	switch interact.ApplicationCommandData().Name {
	case "play":
		message = bot.play(interact)
	case "stop":
		message = bot.stop(interact)
	case "skip":
		message = bot.skip(interact)
	case "queue":
		message = bot.queue(interact)
	}
	_, err = bot.InteractionResponseEdit(bot.State.User.ID, interact.Interaction, &discordgo.WebhookEdit{
		Content: message,
	})
	if err != nil {
		fmt.Println("Error while updating interaction response: ", err)
		return
	}
}

func (bot *Bot) play(interact *discordgo.InteractionCreate) string {

	if bot.musicPlayers[interact.GuildID] == nil {
		bot.addMusicPlayer(interact)
	}
	musicPlayer := bot.musicPlayers[interact.GuildID]

	url := interact.ApplicationCommandData().Options[0].StringValue()
	request, err := player.NewRequest(url, func(title string) { bot.ChannelMessageSend(interact.ChannelID, "**Now Playing:** `"+title+"`") })
	if err != nil {
		message := "Could not add request to queue: There was an error retrieving the [video](" + url + ")."
		return message
	}
	musicPlayer.AddToQueue(request)
	message := "*Added to Queue:* [`" + request.Title + "`](" + url + ")"
	return message
}

func (bot *Bot) stop(interact *discordgo.InteractionCreate) string {

	if bot.musicPlayers[interact.GuildID] == nil {
		message := "I'm not playing any music right now!"
		return message
	}
	musicPlayer := bot.musicPlayers[interact.GuildID]
	musicPlayer.Stop <- true
	message := "Music stopped"
	return message
}

func (bot *Bot) skip(interact *discordgo.InteractionCreate) string {

	if bot.musicPlayers[interact.GuildID] == nil {
		message := "I'm not playing any music right now!"
		return message
	}
	musicPlayer := bot.musicPlayers[interact.GuildID]
	musicPlayer.Next <- true
	message := "Skipped song"
	return message
}

func (bot *Bot) queue(interact *discordgo.InteractionCreate) string {

	if bot.musicPlayers[interact.GuildID] == nil {
		message := "I'm not playing any music right now!"
		return message
	}
	musicPlayer := bot.musicPlayers[interact.GuildID]

	go func() {

		_, err := bot.ChannelMessageSend(interact.ChannelID, "1. **`"+musicPlayer.CurrentSong.Title+"`** - *Now Playing*")
		if err != nil {
			fmt.Println("Error sending channel message: ", err)
		}
		for i, song := range musicPlayer.Queue {
			_, err := bot.ChannelMessageSend(interact.ChannelID, strconv.Itoa(i+2)+". `"+song.Title+"`")
			if err != nil {
				fmt.Println("Error sending channel message: ", err)
			}
		}

	}()

	message := "__Song Queue__"
	return message
}

func (bot *Bot) addMusicPlayer(interact *discordgo.InteractionCreate) {
	musicPlayer := player.NewMusicPlayer()
	bot.musicPlayers[interact.GuildID] = musicPlayer
	invokingMemberChannel, err := bot.State.VoiceState(interact.GuildID, interact.Member.User.ID)
	if err != nil {
		_, err := bot.InteractionResponseEdit(bot.State.User.ID, interact.Interaction, &discordgo.WebhookEdit{
			Content: "You are not currently joined to a voice channel! Please join a voice channel to play music.",
		})
		if err != nil {
			fmt.Println("Error while updating interaction response: ", err)
			return
		}
		return
	}
	//player, err := bot.newMusicPlayer(invokingMemberChannel.GuildID, invokingMemberChannel.ChannelID, false, false)
	go func() {
		musicPlayer.VoiceConnection, err = bot.ChannelVoiceJoin(invokingMemberChannel.GuildID, invokingMemberChannel.ChannelID, false, false)
		if err != nil {
			fmt.Println("Error while joining channel: ", err)
			return
		}
		musicPlayer.LogLevel = discordgo.LogInformational
		musicPlayer.Run()
		delete(bot.musicPlayers, musicPlayer.GuildID)
	}()
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
