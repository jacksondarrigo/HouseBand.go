package houseband

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func (bot *Bot) play(interact *discordgo.InteractionCreate) {
	//
	// Send deferred interaction response
	//
	err := bot.InteractionRespond(interact.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		fmt.Println("Error while sending interaction response: ", err)
		return
	}

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
	request := newRequest(video, interact.ChannelID, bot.ChannelMessageSend)
	player.queue <- request

	//
	// Updated deferred response
	//
	_, err = bot.InteractionResponseEdit(bot.State.User.ID, interact.Interaction, &discordgo.WebhookEdit{
		Content: interact.Member.User.String() + " requested: [" + url + "](" + url + ")",
	})
	if err != nil {
		fmt.Println("Error while updating interaction response: ", err)
		return
	}
	_, err = bot.ChannelMessageSend(interact.ChannelID, "*Added to Queue:* `"+request.Title+"`")
	if err != nil {
		fmt.Println("Error sending channel message: ", err)
	}

}

func (bot *Bot) stop(interact *discordgo.InteractionCreate) {
	//
	// Send deferred interaction response
	//
	err := bot.InteractionRespond(interact.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		fmt.Println("Error while sending interaction response: ", err)
		return
	}

	if bot.musicPlayers[interact.GuildID] == nil {
		_, err = bot.InteractionResponseEdit(bot.State.User.ID, interact.Interaction, &discordgo.WebhookEdit{
			Content: "I'm not playing any music right now!",
		})
		if err != nil {
			fmt.Println("Error while updating interaction response: ", err)
		}
		return
	}
	player := bot.musicPlayers[interact.GuildID]
	player.stop <- true
	_, err = bot.InteractionResponseEdit(bot.State.User.ID, interact.Interaction, &discordgo.WebhookEdit{
		Content: "Stopped playing song",
	})
	if err != nil {
		fmt.Println("Error while updating interaction response: ", err)
		return
	}
}

func (bot *Bot) skip(interact *discordgo.InteractionCreate) {
	//
	// Send deferred interaction response
	//
	err := bot.InteractionRespond(interact.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		fmt.Println("Error while sending interaction response: ", err)
		return
	}

	if bot.musicPlayers[interact.GuildID] == nil {
		_, err = bot.InteractionResponseEdit(bot.State.User.ID, interact.Interaction, &discordgo.WebhookEdit{
			Content: "I'm not playing any music right now!",
		})
		if err != nil {
			fmt.Println("Error while updating interaction response: ", err)
		}
		return
	}
	player := bot.musicPlayers[interact.GuildID]
	player.next <- true
	_, err = bot.InteractionResponseEdit(bot.State.User.ID, interact.Interaction, &discordgo.WebhookEdit{
		Content: "Skipped song",
	})
	if err != nil {
		fmt.Println("Error while updating interaction response: ", err)
		return
	}
}
