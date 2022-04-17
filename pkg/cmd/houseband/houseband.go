package houseband

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	*discordgo.Session
	musicPlayers map[string]*musicPlayer
}

func NewBot(token string) *Bot {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return nil
	}
	return &Bot{session, make(map[string]*musicPlayer)}
}

func (bot *Bot) Run() {
	bot.AddHandler(bot.ready)
	bot.AddHandler(bot.interactionHandler)
	bot.Identify.Intents = discordgo.IntentsAllWithoutPrivileged
	err := bot.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
		return
	}
	fmt.Println("HouseBandTest is now running.  Press CTRL-C to exit.")
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-signals
	// Remove all existing commands on exit - prevents users from issuing a command when the bot is unavailable
	bot.deleteCommands()
	bot.Close()
}

func (bot *Bot) deleteCommands() {
	commands, err := bot.ApplicationCommands(bot.State.User.ID, "485945698953723905")
	if err != nil {
		fmt.Println("Error while getting commands: ", err)
		return
	}
	for _, command := range commands {
		err = bot.ApplicationCommandDelete(bot.State.User.ID, "485945698953723905", command.ID)
		if err != nil {
			fmt.Println("Error while clearing commands: ", err)
		}
	}
}

func (bot *Bot) createCommands() {
	var play_command = &discordgo.ApplicationCommand{
		Name:        "play",
		Description: "Oh.. oh, song? You want to sing a song? You were excited about singing a song? GOOOOOOOOOOD!",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "url",
				Description: "URL to song",
				Required:    true,
			},
		},
	}
	_, err := bot.ApplicationCommandCreate(bot.State.User.ID, "485945698953723905", play_command)
	if err != nil {
		fmt.Println("Error while registering commands: ", err)
	}
}

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
		invokingMemberChannel, err := bot.State.VoiceState(interact.GuildID, interact.Member.User.ID)
		if err != nil {
			fmt.Println("Error while getting user channel: ", err)
			return
		}
		bot.musicPlayers[interact.GuildID] = bot.newMusicPlayer(invokingMemberChannel)
	}
	player := bot.musicPlayers[interact.GuildID]

	//
	// Create and queue songRequest from URL provided by user
	//
	url := interact.ApplicationCommandData().Options[0].StringValue()
	request := newSongRequest(url, interact.ChannelID)
	player.queue <- request

	//
	// Updated deferred response
	//
	_, err = bot.InteractionResponseEdit(bot.State.User.ID, interact.Interaction, &discordgo.WebhookEdit{
		Content: "*Added to Queue:* `" + url + "`",
	})
	if err != nil {
		fmt.Println("Error while updating interaction response: ", err)
		return
	}
	go func() {
		nowPlaying := <-player.nowPlaying
		_, err := bot.ChannelMessageSend(nowPlaying.requestChannelID, "**Now Playing:** `"+nowPlaying.title+"`")
		if err != nil {
			fmt.Println("Error while sending channel message: ", err)
		}

	}()
	go func() {
		<-player.stop
		delete(bot.musicPlayers, player.GuildID)
	}()
}
