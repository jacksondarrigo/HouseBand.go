package bot

import (
	"log"
	"strconv"
	"strings"

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
}

func (bot *Bot) commandHandler(interaction *discordgo.InteractionCreate) {
	err := bot.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Println("Error: Cannot send interaction response: ", err)
		return
	}
	if bot.LogLevel > 1 {
		var builder strings.Builder
		builder.WriteString("[HB] " + interaction.Member.User.Username + " used '" + interaction.ApplicationCommandData().Name + "' command")
		if len(interaction.ApplicationCommandData().Options) > 0 {
			builder.WriteString(" with query '" + interaction.ApplicationCommandData().Options[0].StringValue() + "'")
		}
		log.Println(builder.String())
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
	close(interactionResponse)
	_, err = bot.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: &response,
	})
	if err != nil {
		log.Println("Error: Cannot update interaction response: ", err)
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

	// Get song query
	var query string = interact.ApplicationCommandData().Options[0].StringValue()

	// **
	// ** Note:
	// ** 	Logic for determining whether the music player exists can likely be moved here,
	// ** allowing us to improve clarity by calling getPlayer() and createPlayer() together
	// **

	// Get guild music player if it exists; if it does not exist, create one
	musicPlayer, err := bot.getPlayer(interact, voiceChannel)
	if musicPlayer == nil {
		response := "Error: Cannot join voice channel: " + err.Error()
		interactionResponse <- response
		return
	}

	// Generate request object(s)
	songRequests := make(chan *request.Request)
	go request.Generate(query, interact.ChannelID, songRequests)

	// Receive request object(s) and add them to queue
	for {
		songRequest, ok := <-songRequests
		if !ok {
			break
		}
		musicPlayer.AddToQueue(songRequest)
	}
}

func (bot *Bot) getPlayer(interact *discordgo.InteractionCreate, voiceChannel *discordgo.VoiceState) (*player.MusicPlayer, error) {
	// Get the active music player for the guild. If no such player exists, create one
	var musicPlayer *player.MusicPlayer
	musicPlayer, ok := bot.musicPlayers[interact.GuildID]
	if !ok {
		var err error
		musicPlayer, err = bot.createPlayer(voiceChannel)
		if err != nil {
			return nil, err
		}
		bot.musicPlayers[interact.GuildID] = musicPlayer
	}
	return musicPlayer, nil
}

func (bot *Bot) createPlayer(voiceChannel *discordgo.VoiceState) (*player.MusicPlayer, error) {
	var err error
	var vc *discordgo.VoiceConnection
	for attempts := 0; attempts < 3; attempts++ {
		vc, err = bot.ChannelVoiceJoin(voiceChannel.GuildID, voiceChannel.ChannelID, false, false)
		if err != nil {
			continue
		} else {
			musicPlayer := player.New(vc)
			go bot.routeMessages(musicPlayer)
			return musicPlayer, nil
		}
	}
	return nil, err
}

func (bot *Bot) routeMessages(musicPlayer *player.MusicPlayer) {
	// Route messages from the music player to a Discord channel (such as "Added to Queue" and "Now Playing" messages, as well as error messages)
	for {
		message, ok := <-musicPlayer.Messages
		if !ok {
			break
		}
		_, err := bot.ChannelMessageSend(message.ChannelId, message.Content)
		if err != nil {
			log.Println("Error: Cannot send message to channel: ", err)
		}
	}

	delete(bot.musicPlayers, musicPlayer.GuildID)
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
		message = "Skipped `" + musicPlayer.CurrentSong.Title + "`"
		musicPlayer.Next <- true
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
