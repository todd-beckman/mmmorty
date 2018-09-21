package evalplugin

import (
	"encoding/json"

	"github.com/todd-beckman/mmmorty"
)

const (
	leaveGuild = "leave"
)

// EvalPlugin a
type EvalPlugin struct {
	bot *mmmorty.Bot
}

// Help a
func (e *EvalPlugin) Help(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage) []string {
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
}
