package main

import (
	"flag"
	"fmt"

	"code.darrigo.dev/houseband/pkg/cmd/houseband"
)

func main() {

	var token *string = flag.String("t", "", "Discord API Token")
	flag.Parse()
	if *token == "" {
		fmt.Println("No token provided. Please use the '-t' option to set your Discord API Token.")
		return
	}
	bot := houseband.NewBot(*token)
	bot.Run()
}
