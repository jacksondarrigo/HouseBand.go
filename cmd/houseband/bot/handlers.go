package bot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jacksondarrigo/HouseBand.go/cmd/houseband/player"
	"github.com/jacksondarrigo/HouseBand.go/cmd/houseband/request"
)

func (bot *Bot) onReady(session *discordgo.Session, event *discordgo.Ready) {
	bot.createCommands()
}

func (bot *Bot) onReadyReset(session *discordgo.Session, event *discordgo.Ready) {
	bot.deleteCommands()
	bot.createCommands()
}

func (bot *Bot) createCommands() {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "play",
			Description: "Play a song",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "query",
					Description: "Youtube search query, Youtube video ID, or URL to song",
					Required:    true,
				},
			},
		},
		{
			Name:        "stop",
			Description: "Stop playing music and disconnect",
		},
		{
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
	now := time.Now()
	fmt.Println(now, interact.ApplicationCommandData().Name, "command used by", interact.Member.User.Username)
	messageChannel := make(chan string)
	switch interact.ApplicationCommandData().Name {
	case "play":
		go bot.play(interact, messageChannel)
	case "stop":
		go bot.stop(interact, messageChannel)
	case "skip":
		go bot.skip(interact, messageChannel)
	case "queue":
		go bot.queue(interact, messageChannel)
	}
	message := <-messageChannel
	_, err = bot.InteractionResponseEdit(bot.State.User.ID, interact.Interaction, &discordgo.WebhookEdit{
		Content: message,
	})
	if err != nil {
		fmt.Println("Error while updating interaction response: ", err)
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

func (bot *Bot) play(interact *discordgo.InteractionCreate, messageChannel chan<- string) {

	var query string = interact.ApplicationCommandData().Options[0].StringValue()
	var nowPlaying chan bool = make(chan bool)

	invokingMemberChannel, err := bot.State.VoiceState(interact.GuildID, interact.Member.User.ID)
	if err != nil {
		message := "You are not currently joined to a voice channel! Please join a voice channel to play music."
		messageChannel <- message
		return
	}

	req, err := request.New(query, nowPlaying)
	if err != nil {
		message := "Could not add request to queue: " + err.Error()
		messageChannel <- message
		return
	}
	message := interact.Member.User.Username + " requested: [`" + req.Title + "`](" + req.ReqURL + ")"
	messageChannel <- message

	bot.mu.Lock()
	defer bot.mu.Unlock()
	if bot.musicPlayers[interact.GuildID] == nil {
		bot.musicPlayers[interact.GuildID] = player.New()
	}
	musicPlayer := bot.musicPlayers[interact.GuildID]
	musicPlayer.AddToQueue(req)
	go func() {
		nowPlaying := <-nowPlaying
		if nowPlaying {
			bot.ChannelMessageSend(interact.ChannelID, "**Now Playing:** `"+req.Title+"`")
		} else {
			bot.ChannelMessageSend(interact.ChannelID, "**Error Playing:** `"+req.Title+"`; *skipping song*")
		}
	}()
	bot.ChannelMessageSend(interact.ChannelID, "*Added to Queue:* `"+req.Title+"`")
	if !musicPlayer.Started {
		go bot.startPlayer(interact, invokingMemberChannel)
	}
	// message := "*Added to Queue:* [`" + req.Title + "`](" + req.ReqURL + ")"

}

func (bot *Bot) stop(interact *discordgo.InteractionCreate, messageChannel chan<- string) {

	if bot.musicPlayers[interact.GuildID] == nil {
		message := "I'm not playing any music right now!"
		messageChannel <- message
		return
	}
	musicPlayer := bot.musicPlayers[interact.GuildID]

	musicPlayer.Stop <- true
	message := "Music stopped"
	messageChannel <- message
}

func (bot *Bot) skip(interact *discordgo.InteractionCreate, messageChannel chan<- string) {

	if bot.musicPlayers[interact.GuildID] == nil {
		message := "I'm not playing any music right now!"
		messageChannel <- message
		return
	}
	musicPlayer := bot.musicPlayers[interact.GuildID]

	musicPlayer.Next <- true
	message := "Skipped song"
	messageChannel <- message
}

func (bot *Bot) queue(interact *discordgo.InteractionCreate, messageChannel chan<- string) {

	if bot.musicPlayers[interact.GuildID] == nil {
		message := "I'm not playing any music right now!"
		messageChannel <- message
		return
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
	messageChannel <- message
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
