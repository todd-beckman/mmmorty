package evalplugin

import (
	"encoding/json"
	"fmt"

	"github.com/todd-beckman/mmmorty"
)

const (
	eval       = "eval"
	leaveGuild = "leave"
)

// EvalPlugin a
type EvalPlugin struct {
	bot *mmmorty.Bot
}

// Help a
func (e *EvalPlugin) Help(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage, detail bool) []string {
	return []string{}
}

// Save a
func (e *EvalPlugin) Save() ([]byte, error) {
	return json.Marshal(e)
}

// Name a
func (e *EvalPlugin) Name() string {
	return "Eval"
}

// New a
func New() mmmorty.Plugin {
	return &EvalPlugin{}
}

// Load a
func (e *EvalPlugin) Load(bot *mmmorty.Bot, service mmmorty.Discord, data []byte) error {
	return nil
}

// Message a
func (e *EvalPlugin) Message(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage) {
	defer bot.MessageRecover(service, message.Channel())

	if service.IsMe(message) || service.IsPrivate(message) || message.UserID() != service.OwnerUserID {
		return
	}

	if !mmmorty.MatchesCommand(service, eval, message) {
		return
	}

	requester := fmt.Sprintf("<@%s>", message.UserID())

	channelID := message.Channel()
	discordChannel, err := service.Channel(channelID)
	if err != nil {
		reply := fmt.Sprintf("Uh, %s, something went figuring out your server.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	_, parts := mmmorty.ParseCommand(service, message)

	if len(parts) == 0 {
		return
	}

	command := parts[1]

	if command == "leave" {
		guildID := discordChannel.GuildID

		service.GuildLeave(guildID)

		return
	}
}
