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
	pickCommand  = "pick between"
	pickTemplate = "Uh, I'll go with this one: %s"
)

type PickPlugin struct {
	bot *mmmorty.Bot
}

func (p *PickPlugin) Help(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message, detailed bool) []string {
	return mmmorty.CommandHelp(service, pickCommand, "pick between <option> or <option> (or ...)",
		"asks Morty to pick something for you")
}

func (p *PickPlugin) Load(bot *mmmorty.Bot, service mmmorty.Service, data []byte) error {
	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
			return err
		}
	}

	return nil
}

func (p *PickPlugin) Message(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message) {
	defer mmmorty.MessageRecover()

	if service.Name() != mmmorty.DiscordServiceName {
		return
	}

	if service.IsMe(message) {
		return
	}

	if mmmorty.MatchesCommand(service, pickCommand, message) {
		p.handlePickCommand(bot, service, message)
	}
}

func (p *PickPlugin) handlePickCommand(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	if strings.Contains(message.Message(), "http") {
		reply := fmt.Sprintf("Uh, %s, I would rather not pick between links.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	_, parts := mmmorty.ParseCommand(service, message)

	options := []string{}
	currentOption := []string{}

	for index, word := range parts[1:] { // first word is 'between' because 'pick between'
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
		reply := fmt.Sprintf("Uh, %s, I didn't get that. Maybe put `or` between option?", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	index := rand.Intn(len(options))
	choice := options[index]
	reply := fmt.Sprintf(pickTemplate, choice)
	service.SendMessage(message.Channel(), reply)
}

func (p *PickPlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

func (p *PickPlugin) Stats(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message) []string {
	return []string{}
}

func (p *PickPlugin) Name() string {
	return "Pick"
}

func New() mmmorty.Plugin {
	return &PickPlugin{}
}
