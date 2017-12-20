package pickplugin

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/todd-beckman/mmmorty"
)

const (
	maxWordCount = 100
	pickCommand  = "choose"
	pickTemplate = "Uh, I'll go with this one: %s"
)

// PickPlugin is a the save structure for this plugin
type PickPlugin struct {
	bot *mmmorty.Bot
}

// Help gets the usage for this plugin
func (p *PickPlugin) Help(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage, detailed bool) []string {
	return mmmorty.CommandHelp(service, pickCommand, "option 1 or option 2 or ...",
		"asks Morty to pick between an arbitrary number of things for you")
}

// Load loads the plugin from the given data
func (p *PickPlugin) Load(bot *mmmorty.Bot, service mmmorty.Discord, data []byte) error {
	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
			return err
		}
	}

	return nil
}

// Message is the command handler for this plugin
func (p *PickPlugin) Message(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage) {
	defer bot.MessageRecover(service, message.Channel())

	if service.IsMe(message) {
		return
	}

	if mmmorty.MatchesCommand(service, pickCommand, message) {
		p.handlePickCommand(bot, service, message)
	}
}

func (p *PickPlugin) handlePickCommand(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	if strings.Contains(message.Message(), "http") {
		reply := fmt.Sprintf("Uh, %s, I would rather not pick between links.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	_, parts := mmmorty.ParseCommand(service, message)

	options := []string{}
	currentOption := []string{}

	for index, word := range parts {
		if index > maxWordCount {
			reply := fmt.Sprintf("Uh, %s, that message is kind of long. Is there any way you can shorten it?", requester)
			service.SendMessage(message.Channel(), reply)
			return
		}
		if word == "or" {
			options = append(options, strings.Join(currentOption, " "))
			currentOption = []string{}
		} else {
			currentOption = append(currentOption, word)
		}
	}

	if len(currentOption) == 0 {
		reply := fmt.Sprintf("Uh, %s, I didn't get that last one.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	options = append(options, strings.Join(currentOption, " "))

	if len(options) < 2 {
		reply := fmt.Sprintf("Uh, %s, I didn't get that. Maybe put `or` between options?", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	index := rand.Intn(len(options))
	choice := options[index]
	reply := fmt.Sprintf(pickTemplate, choice)
	service.SendMessage(message.Channel(), reply)
}

// Save saves the plugin's state to file
func (p *PickPlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

// Name gets the name of the plugin for saving purposes
func (p *PickPlugin) Name() string {
	return "Pick"
}

// New creates a new instance of this plugin
func New() mmmorty.Plugin {
	return &PickPlugin{}
}
