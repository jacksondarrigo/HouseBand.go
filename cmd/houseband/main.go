package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jacksondarrigo/HouseBand.go/cmd/houseband/bot"
)

func main() {

	tFlag := flag.String("t", "", "Discord API Token")
	rFlag := flag.Bool("r", false, "Reset all bot commands")
	flag.Parse()
	token := *tFlag
	resetCommands := *rFlag

	if token == "" {
		token = os.Getenv("DISCORD_TOKEN")
		if token == "" {
			fmt.Println("No token provided. Please set DISCORD_TOKEN environment variable, or use '-t' option to set your Discord API token.")
			return
		}
	}

	// var token string

	// env := os.Getenv("DISCORD_TOKEN")
	// if env == "" {
	// 	tokenFlag := flag.String("t", "", "Discord API Token")
	// 	flag.Parse()
	// 	if *tokenFlag == "" {
	// 		fmt.Println("No token provided. Please use the '-t' option to set your Discord API Token.")
	// 		return
	// 	}
	// 	token = *tokenFlag
	// } else {
	// 	token = env
	// }
	houseband := bot.New(token)
	houseband.Run(resetCommands)
}
