package bot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jacksondarrigo/HouseBand.go/cmd/houseband/player"
	"github.com/jacksondarrigo/HouseBand.go/cmd/houseband/request"
)

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

func (bot *Bot) startPlayer(interact *discordgo.InteractionCreate, channel *discordgo.VoiceState) {
	musicPlayer := bot.musicPlayers[interact.GuildID]
	musicPlayer.Started = true
	var err error
	musicPlayer.VoiceConnection, err = bot.ChannelVoiceJoin(channel.GuildID, channel.ChannelID, false, false)
	if err != nil {
		bot.ChannelMessageSend(interact.ChannelID, "Error while joining voice channel: "+err.Error())
	} else {
		musicPlayer.Start()
	}
	musicPlayer.Disconnect()
	musicPlayer.Started = false
	delete(bot.musicPlayers, interact.GuildID)
}

func (bot *Bot) play(interact *discordgo.InteractionCreate) string {

	var url string = interact.ApplicationCommandData().Options[0].StringValue()
	var nowPlaying chan bool = make(chan bool)

	invokingMemberChannel, err := bot.State.VoiceState(interact.GuildID, interact.Member.User.ID)
	if err != nil {
		message := "You are not currently joined to a voice channel! Please join a voice channel to play music."
		return message
	}

	req, err := request.New(url, nowPlaying)
	if err != nil {
		message := "Could not add request to queue: " + err.Error()
		return message
	}

	bot.mu.Lock()
	if bot.musicPlayers[interact.GuildID] == nil {
		bot.musicPlayers[interact.GuildID] = player.New()
	}
	musicPlayer := bot.musicPlayers[interact.GuildID]
	musicPlayer.AddToQueue(req)
	bot.mu.Unlock()

	go func() {
		nowPlaying := <-nowPlaying
		if nowPlaying {
			bot.ChannelMessageSend(interact.ChannelID, "**Now Playing:** `"+req.Title+"`")
		} else {
			bot.ChannelMessageSend(interact.ChannelID, "**Error Playing:** `"+req.Title+"`; *skipping song*")
		}
	}()
	bot.mu.Lock()
	if !musicPlayer.Started {
		go bot.startPlayer(interact, invokingMemberChannel)
	}
	bot.mu.Unlock()
	message := "*Added to Queue:* [`" + req.Title + "`](" + url + ")"
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

		var builder strings.Builder

		builder.WriteString("`1.` **`" + musicPlayer.CurrentSong.Title + "`** - *Now Playing*\n")
		for i := 0; i < len(musicPlayer.Queue); i++ {
			builder.WriteString("`" + strconv.Itoa(i+2) + ".` `" + musicPlayer.Queue[i].Title + "`\n")
		}
		_, err := bot.ChannelMessageSend(interact.ChannelID, builder.String())
		if err != nil {
			fmt.Println("Error sending channel message: ", err)
		}

	}()

	message := "__Song Queue__"
	return message
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

// func (bot *Bot) onVoiceJoin(session *discordgo.Session, event *discordgo.VoiceServerUpdate) {
// 	bot.createCommands()
// }
