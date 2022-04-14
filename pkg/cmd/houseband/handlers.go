package houseband

import "github.com/bwmarrin/discordgo"

func (bot *Bot) ready(session *discordgo.Session, event *discordgo.Ready) {
	bot.createCommands()
}

func (bot *Bot) interactionHandler(session *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		bot.commandHandler(i)
	case discordgo.InteractionMessageComponent:
		bot.componentHandler(i)
	}
}

func (bot *Bot) commandHandler(i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "play":
		bot.play(i)
	}
}

// TODO: Implement component handler
func (bot *Bot) componentHandler(i *discordgo.InteractionCreate) {
	return
}
