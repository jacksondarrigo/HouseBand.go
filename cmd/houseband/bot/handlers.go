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

func (bot *Bot) commandHandler(interaction *discordgo.InteractionCreate) {
	err := bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		fmt.Println("Error: Cannot send interaction response: ", err)
		return
	}
	if bot.LogLevel > 1 {
		var builder strings.Builder
		builder.WriteString(time.Now().String() + " [HB] " + interaction.Member.User.Username + " used " + interaction.ApplicationCommandData().Name + " command")
		if interaction.ApplicationCommandData().Options[0] != nil {
			builder.WriteString(" with query " + interaction.ApplicationCommandData().Options[0].StringValue())
		}
		fmt.Println(builder.String())
	}
	interactionResponse := make(chan string)
	switch interaction.ApplicationCommandData().Name {
	case "play", "test_play":
		go bot.play(interaction, interactionResponse)
	case "stop", "test_stop":
		go bot.stop(interaction, interactionResponse)
	case "skip", "test_skip":
		go bot.skip(interaction, interactionResponse)
	case "queue", "test_queue":
		go bot.queue(interaction, interactionResponse)
	}
	response := <-interactionResponse
	_, err = bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: &response,
	})
	if err != nil {
		fmt.Println("Error: Cannot update interaction response: ", err)
	}
}

func (bot *Bot) play(interact *discordgo.InteractionCreate, interactionResponse chan<- string) {

	// Check to see if user is connected to a voice channel in the guild where the command was issued
	voiceChannel, err := bot.State.VoiceState(interact.GuildID, interact.Member.User.ID)
	if err != nil {
		response := "You are not currently joined to a voice channel! Please join a voice channel to play music."
		interactionResponse <- response
		return
	}

	// Get song query and create request object
	var query string = interact.ApplicationCommandData().Options[0].StringValue()
	req, err := request.New(query, interact.ChannelID)
	if err != nil {
		response := "Error: Could not complete request for '" + query + "': " + err.Error()
		interactionResponse <- response
		return
	}
	response := interact.Member.User.Username + " requested: [`" + req.Title + "`](" + req.ReqURL + ")"
	interactionResponse <- response

	// Get/create music player and add request to queue
	musicPlayer := bot.getPlayer(interact, voiceChannel)
	musicPlayer.AddToQueue(req)

}

func (bot *Bot) getPlayer(interact *discordgo.InteractionCreate, voiceChannel *discordgo.VoiceState) *player.MusicPlayer {
	// Get the active music player for the guild. If no such player exists, create one
	musicPlayer, ok := bot.musicPlayers[interact.GuildID]
	if !ok {
		musicPlayer = player.New()
		go bot.startPlayer(musicPlayer, interact, voiceChannel)
		go bot.receiveMessages(musicPlayer)
		bot.musicPlayers[interact.GuildID] = musicPlayer
	}
	return musicPlayer
}

func (bot *Bot) startPlayer(musicPlayer *player.MusicPlayer, interact *discordgo.InteractionCreate, voiceChannel *discordgo.VoiceState) {
	// Connect the music player to a voice channel, then start the main player loop. Cleanup afterwards
	var err error
	musicPlayer.VoiceConnection, err = bot.ChannelVoiceJoin(voiceChannel.GuildID, voiceChannel.ChannelID, false, false)
	if err != nil {
		musicPlayer.Messages <- player.Message{ChannelId: interact.ChannelID, Content: "Error: Cannot join voice channel: " + err.Error()}
	} else {
		musicPlayer.Run()
		musicPlayer.Disconnect()
	}
	close(musicPlayer.Messages)
	delete(bot.musicPlayers, interact.GuildID)
}

func (bot *Bot) receiveMessages(musicPlayer *player.MusicPlayer) {
	// Receive messages from the music player to forward to a Discord channel (such as "Added to Queue" and "Now Playing" messages, as well as error messages)
	for {
		select {
		case message, ok := <-musicPlayer.Messages:
			if !ok {
				return
			}
			_, err := bot.ChannelMessageSend(message.ChannelId, message.Content)
			if err != nil {
				fmt.Println("Error: Cannot send message to channel: ", err)
			}
		}
	}
}

func (bot *Bot) stop(interact *discordgo.InteractionCreate, interactionResponse chan<- string) {

	var message string

	if musicPlayer, ok := bot.musicPlayers[interact.GuildID]; ok {
		musicPlayer.Stop <- true
		message = "Music stopped"
	} else {
		message = "I'm not playing any music right now!"
	}

	interactionResponse <- message
}

func (bot *Bot) skip(interact *discordgo.InteractionCreate, interactionResponse chan<- string) {

	var message string

	if musicPlayer, ok := bot.musicPlayers[interact.GuildID]; ok {
		musicPlayer.Next <- true
		message = "Skipped song"
	} else {
		message = "I'm not playing any music right now!"
	}

	interactionResponse <- message

}

func (bot *Bot) queue(interact *discordgo.InteractionCreate, interactionResponse chan<- string) {

	var message string

	if musicPlayer, ok := bot.musicPlayers[interact.GuildID]; ok {
		go func() {

			var builder strings.Builder

			builder.WriteString("`1.` **`" + musicPlayer.CurrentSong.Title + "`** - *Now Playing*\n")
			for i := 0; i < len(musicPlayer.Queue); i++ {
				builder.WriteString("`" + strconv.Itoa(i+2) + ".` `" + musicPlayer.Queue[i].Title + "`\n")
			}

			musicPlayer.Messages <- player.Message{ChannelId: interact.ChannelID, Content: builder.String()}

		}()

		message = "__Song Queue__"
	} else {
		message = "I'm not playing any music right now!"
	}

	interactionResponse <- message
}
