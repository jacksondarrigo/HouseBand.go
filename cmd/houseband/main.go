package main

import (
	"flag"
	"fmt"

	"github.com/jacksondarrigo/HouseBand.go/cmd/houseband/bot"
)

func main() {

	var token *string = flag.String("t", "", "Discord API Token")
	flag.Parse()
	if *token == "" {
		fmt.Println("No token provided. Please use the '-t' option to set your Discord API Token.")
		return
	}
	bot := bot.NewBot(*token)
	bot.Run()
}
