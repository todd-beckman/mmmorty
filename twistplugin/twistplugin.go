package twistplugin

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/todd-beckman/mmmorty"
)

const (
	addTwistCommand = "add plot twist"
	twistCommand    = "plot twist"
	twistTemplate   = "```\n" +
		"%s\n" +
		"```"

	maxTwistCount = 500
	maxWordCount  = 150
)

type Twist struct {
	Twist string `json: "twist"`
}

type TwistPlugin struct {
	bot    *mmmorty.Bot
	Twists []Twist `json: "twists"`
}

func (p *TwistPlugin) Help(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message, detailed bool) []string {
	help := mmmorty.CommandHelp(service, addTwistCommand, "some plot twist", "adds a plot twist for Morty to remember")
	help = append(help, mmmorty.CommandHelp(service, twistCommand, "", "asks Morty for a plot twist at random.")[0])
	return help
}

func (p *TwistPlugin) Load(bot *mmmorty.Bot, service mmmorty.Service, data []byte) error {
	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
			return err
		}
	}

	return nil
}

func (p *TwistPlugin) Message(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message) {
	defer bot.MessageRecover(service, message.Channel())

	if service.Name() != mmmorty.DiscordServiceName {
		return
	}

	if service.IsMe(message) {
		return
	}

	if mmmorty.MatchesCommand(service, addTwistCommand, message) {
		p.handleAddTwistCommand(bot, service, message)
	} else if mmmorty.MatchesCommand(service, twistCommand, message) {
		p.handleTwistCommand(bot, service, message)
	}
}

func (p *TwistPlugin) handleAddTwistCommand(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	if service.IsPrivate(message) {
		reply := fmt.Sprintf("Uh, %s, I can't add plot twists privately.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	if len(p.Twists) >= maxTwistCount {
		reply := fmt.Sprintf("Uh, %s, I can't remember all these plot twists. Rick might need to help get rid of some.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	if strings.Contains(message.Message(), "http") {
		reply := fmt.Sprintf("Uh, %s, I would rather not remember plot twists with links.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	_, parts := mmmorty.ParseCommand(service, message)

	twistParts := []string{}
	for index, word := range parts[2:] {
		if index > maxWordCount {
			reply := fmt.Sprintf("Uh, %s, that twist is kind of long. Is there any way you can shorten it?", requester)
			service.SendMessage(message.Channel(), reply)
			return
		} else {
			twistParts = append(twistParts, word)
		}
	}

	if len(twistParts) == 0 {
		reply := fmt.Sprintf("Uh, %s, I didn't get that. Be sure to say the plot twist you want me to remember.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	plotTwist := strings.Join(twistParts, " ")

	newTwist := Twist{
		Twist: plotTwist,
	}

	p.Twists = append(p.Twists, newTwist)

	reply := fmt.Sprintf("Ok, %s, you got it! I will try to remember that one.", requester)
	service.SendMessage(message.Channel(), reply)
}

func (p *TwistPlugin) handleTwistCommand(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	twistCount := len(p.Twists)
	if twistCount == 0 {
		reply := fmt.Sprintf("Uh, %s, I don't know any twists yet. Maybe you could add them?", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	index := rand.Intn(twistCount)
	twist := p.Twists[index]
	reply := fmt.Sprintf(twistTemplate, twist.Twist)
	service.SendMessage(message.Channel(), reply)
}

func (p *TwistPlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

func (p *TwistPlugin) Stats(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message) []string {
	return []string{}
}

func (p *TwistPlugin) Name() string {
	return "Twist"
}

func New() mmmorty.Plugin {
	return &TwistPlugin{}
}
