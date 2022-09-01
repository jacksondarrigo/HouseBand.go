package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jacksondarrigo/HouseBand.go/cmd/houseband/bot"
)

func main() {

	defaultTokenValue := ""

	envToken, ok := os.LookupEnv("DISCORD_TOKEN")
	if ok {
		defaultTokenValue = envToken
	}

	tFlag := flag.String("t", defaultTokenValue, "Discord API Token")
	rFlag := flag.Bool("r", false, "Reset all bot commands")
	lFlag := flag.String("l", "ERROR", "Logging level")
	flag.Parse()
	token := *tFlag
	resetCommands := *rFlag
	logLevel := *lFlag

	if token == "" {
		fmt.Println("No token provided. Please set DISCORD_TOKEN environment variable, or use '-t' option to set your Discord API token.")
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
		fmt.Println("Unknown LogLevel. Please set LogLevel to ERROR, WARN, INFO, or DEBUG.")
		return
	}

	houseband.Run(resetCommands)
}
