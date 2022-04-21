package houseband

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func (bot *Bot) deferResponse(interact *discordgo.InteractionCreate) {
	err := bot.InteractionRespond(interact.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		fmt.Println("Error while sending interaction response: ", err)
		return
	}
}

func (bot *Bot) editResponse(interact *discordgo.InteractionCreate, message string) {
	_, err := bot.InteractionResponseEdit(bot.State.User.ID, interact.Interaction, &discordgo.WebhookEdit{
		Content: message,
	})
	if err != nil {
		fmt.Println("Error while updating interaction response: ", err)
		return
	}
}

// func (bot *Bot) getURL() {
// 	url := interact.ApplicationCommandData().Options[0].StringValue()

// 	video, err := bot.youtube.GetVideo(url)
// 	if err != nil {
// 		fmt.Println("Error while getting video: ", err)
// 	}
// 	stream, err := bot.youtube.GetStreamURL(video, video.Formats.FindByItag(251))
// 	if err != nil {
// 		fmt.Println("Error while getting stream URL: ", err)
// 	}
// }

func (bot *Bot) play(interact *discordgo.InteractionCreate) {
	bot.deferResponse(interact)

	//
	// Check for existing musicPlayer, or create one if one doesn't exist
	//
	if bot.musicPlayers[interact.GuildID] == nil {
		player := newMusicPlayer()
		bot.musicPlayers[interact.GuildID] = player
		go func() {
			invokingMemberChannel, err := bot.State.VoiceState(interact.GuildID, interact.Member.User.ID)
			if err != nil {
				fmt.Println("Error while getting user channel: ", err)
				return
			}
			player.VoiceConnection, err = bot.ChannelVoiceJoin(invokingMemberChannel.GuildID, invokingMemberChannel.ChannelID, false, false)
			if err != nil {
				fmt.Println("Error while joining channel: ", err)
			}
			player.run()
			delete(bot.musicPlayers, player.GuildID)
		}()
	}
	player := bot.musicPlayers[interact.GuildID]

	//
	// Create and queue request from URL provided by user
	//
	url := interact.ApplicationCommandData().Options[0].StringValue()

	video, err := bot.youtube.GetVideo(url)
	if err != nil {
		fmt.Println("Error while getting video: ", err)
	}
	stream, err := bot.youtube.GetStreamURL(video, video.Formats.FindByItag(251))
	if err != nil {
		fmt.Println("Error while getting stream URL: ", err)
	}
	request := newRequest(video, stream, interact.ChannelID, bot.ChannelMessageSend)
	player.queue <- request

	_, err = bot.ChannelMessageSend(interact.ChannelID, "*Added to Queue:* `"+request.Title+"`")
	if err != nil {
		fmt.Println("Error sending channel message: ", err)
	}

	bot.editResponse(interact, interact.Member.User.String()+" requested: ["+url+"]("+url+")")

}

func (bot *Bot) stop(interact *discordgo.InteractionCreate) {
	bot.deferResponse(interact)

	if bot.musicPlayers[interact.GuildID] == nil {
		bot.editResponse(interact, "I'm not playing any music right now!")
		return
	}
	player := bot.musicPlayers[interact.GuildID]
	player.stop <- true

	bot.editResponse(interact, "Stopped playing song")
}

func (bot *Bot) skip(interact *discordgo.InteractionCreate) {
	bot.deferResponse(interact)

	if bot.musicPlayers[interact.GuildID] == nil {
		bot.editResponse(interact, "I'm not playing any music right now!")
		return
	}
	player := bot.musicPlayers[interact.GuildID]
	player.next <- true

	bot.editResponse(interact, "Skipped song")
}
