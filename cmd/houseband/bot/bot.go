package bot

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/jacksondarrigo/HouseBand.go/cmd/houseband/player"
)

type Bot struct {
	*discordgo.Session
	musicPlayers map[string]*player.MusicPlayer
}

func New(token string) *Bot {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error: Cannot create Discord session: ", err)
		return nil
	}
	return &Bot{session, make(map[string]*player.MusicPlayer)}
}

func (bot *Bot) Connect() {
	bot.AddHandler(bot.interactionHandler)
	bot.Identify.Intents = discordgo.IntentsAllWithoutPrivileged
	err := bot.Open()
	if err != nil {
		fmt.Println("Error: Cannot open Discord session: ", err)
		return
	}

}

func (bot *Bot) Wait() {
	defer bot.Close()
	fmt.Println("HouseBandTest is now running.  Press CTRL-C to exit.")
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-signals
}

func (bot *Bot) RegisterCommands(commands []*discordgo.ApplicationCommand) {
	for _, command := range commands {
		_, err := bot.ApplicationCommandCreate(bot.State.User.ID, "", command)
		if err != nil {
			fmt.Println("Error: Cannot create commands: ", err)
		}
	}
}

func (bot *Bot) DeleteCommands(commands []*discordgo.ApplicationCommand) {
	for _, command := range commands {
		err := bot.ApplicationCommandDelete(bot.State.User.ID, "", command.ID)
		if err != nil {
			fmt.Println("Error: Cannot delete commands: ", err)
		}
	}
}
