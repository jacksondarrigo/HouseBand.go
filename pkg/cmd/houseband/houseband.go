package houseband

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/kkdai/youtube/v2"
)

type Bot struct {
	*discordgo.Session
	youtube      youtube.Client
	musicPlayers map[string]*musicPlayer
}

func NewBot(token string) *Bot {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return nil
	}
	return &Bot{session, youtube.Client{}, make(map[string]*musicPlayer)}
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
	commands := []*discordgo.ApplicationCommand{
		{
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
		},
		{
			Name:        "stop",
			Description: "What are you doing? Don't do that... I COMMAND YOU TO STOP.",
		},
		{
			Name:        "skip",
			Description: "... I didn't tell you what to- You're skipping a line, dude.",
		},
	}
	for _, command := range commands {
		_, err := bot.ApplicationCommandCreate(bot.State.User.ID, "485945698953723905", command)
		if err != nil {
			fmt.Println("Error while registering commands: ", err)
		}
	}
}
