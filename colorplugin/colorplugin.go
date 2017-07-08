package colorplugin

import (
	"encoding/json"
	"fmt"
	"log"

	//	"github.com/dustin/go-humanize"
	"github.com/todd-beckman/mmmorty"
)

const colorCommand = "color me"

type ColorPlugin struct {
	bot *mmmorty.Bot
}

func (p *ColorPlugin) Help(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message, detailed bool) []string {
	help := mmmorty.CommandHelp(service, colorCommand, "<color>", "assigns the desired color if it is avialable")
	return help
}

func (p *ColorPlugin) Load(bot *mmmorty.Bot, service mmmorty.Service, data []byte) error {
	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
		}
	}

	return nil
}

func (p *ColorPlugin) Message(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message) {
	defer mmmorty.MessageRecover()

	if service.IsMe(message) {
		return
	}

	if !mmmorty.MatchesCommand(service, "color me", message) {
		return
	}

	_, parts := mmmorty.ParseCommand(service, message)

	if len(parts) != 2 {
		log.Println(fmt.Sprintf("color me message improperly formatted: %s", message))
		return
	}

	color := parts[1]

	requester := message.UserName()
	if service.Name() == mmmorty.DiscordServiceName {
		requester = fmt.Sprintf("<@%s>", message.UserID())
	}

	reply := fmt.Sprintf("Sorry %s, you want to be %s but I can't do that yet", requester, color)
	log.Println(reply)
	service.SendMessage(message.Channel(), reply)
}

// Save will save plugin state to a byte array.
func (p *ColorPlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

// Stats will return the stats for a plugin.
func (p *ColorPlugin) Stats(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message) []string {
	return []string{}
}

// Name returns the name of the plugin.
func (p *ColorPlugin) Name() string {
	return "Reminder"
}

// New will create a new Reminder plugin.
func New() mmmorty.Plugin {
	return &ColorPlugin{}
}
