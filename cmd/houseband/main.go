package main

import (
	"flag"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/jacksondarrigo/HouseBand.go/cmd/houseband/bot"
)

func main() {

	defaultTokenValue := ""
	envToken, ok := os.LookupEnv("DISCORD_TOKEN")
	if ok {
		defaultTokenValue = envToken
	}

	tokenFlag := flag.String("t", defaultTokenValue, "Discord API Token")
	logFlag := flag.String("l", "ERROR", "Log level")
	testFlag := flag.Bool("m", false, "Test mode")
	createFlag := flag.Bool("c", false, "Create/register/update new bot commands")
	deleteFlag := flag.Bool("d", false, "Delete bot commands")
	flag.Parse()

	token := *tokenFlag
	logLevel := *logFlag
	testMode := *testFlag
	createCommands := *createFlag
	deleteCommands := *deleteFlag

	if token == "" {
		log.Println("Error: No token provided. Please set DISCORD_TOKEN environment variable, or use '-t' option to set your Discord API token.")
		return
	}

	houseband := bot.New(token)

	switch logLevel {
	case "ERROR":
		houseband.LogLevel = 0
	case "WARN":
		houseband.LogLevel = 1
	case "INFO":
		houseband.LogLevel = 2
	case "DEBUG":
		houseband.LogLevel = 3
	default:
		log.Println("Error: Unknown log level. Please set log level to one of ERROR, WARN, INFO, or DEBUG.")
		return
	}
	var commands []*discordgo.ApplicationCommand
	if !testMode {
		commands = []*discordgo.ApplicationCommand{
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
	} else {
		commands = []*discordgo.ApplicationCommand{
			{
				Name:        "test_play",
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
				Name:        "test_stop",
				Description: "Stop playing music and disconnect",
			},
			{
				Name:        "test_skip",
				Description: "Skip the current song in queue",
			},
			{
				Name:        "test_queue",
				Description: "List all songs in queue",
			},
		}
	}

	houseband.Connect()

	if createCommands {
		houseband.RegisterCommands(commands)
		houseband.Close()
		return
	}

	if deleteCommands {
		commands, err := houseband.ApplicationCommands(houseband.State.User.ID, "")
		if err != nil {
			log.Println("Error: cannot retrieve application commands:", err.Error())
		} else {
			houseband.DeleteCommands(commands)
		}
		houseband.Close()
		return
	}

	houseband.Wait()
}
