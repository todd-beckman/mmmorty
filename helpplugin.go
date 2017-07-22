package mmmorty

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
)

const helpCommand = "help"

type helpPlugin struct {
}

// Name returns the name of the service.
func (p *helpPlugin) Name() string {
	return "Help"
}

// Help returns a list of help strings that are printed when the user requests them.
func (p *helpPlugin) Help(bot *Bot, service Service, message Message, detailed bool) []string {
	privs := service.SupportsPrivateMessages() && !service.IsPrivate(message) && service.IsModerator(message)
	if detailed && !privs {
		return nil
	}

	commands := []string{}

	for _, plugin := range bot.Services[service.Name()].Plugins {
		hasDetailed := false

		if plugin == p {
			hasDetailed = privs
		} else {
			t := plugin.Help(bot, service, message, true)
			hasDetailed = t != nil && len(t) > 0
		}

		if hasDetailed {
			commands = append(commands, strings.ToLower(plugin.Name()))
		}
	}

	sort.Strings(commands)

	help := []string{}

	if len(commands) > 0 {
		help = append(help, CommandHelp(service, helpCommand, "[topic]", fmt.Sprintf("Returns this information."))[0])
	}

	return help
}

func (p *helpPlugin) Message(bot *Bot, service Service, message Message) {
	if !service.IsMe(message) {
		if MatchesCommand(service, "help", message) || MatchesCommand(service, "command", message) {

			_, parts := ParseCommand(service, message)

			help := []string{}

			for _, plugin := range bot.Services[service.Name()].Plugins {
				h := plugin.Help(bot, service, message, false)
				if h != nil && len(h) > 0 {
					help = append(help, h...)
				}
			}

			if len(parts) == 0 {
				sort.Strings(help)
				if service.SupportsPrivateMessages() {
					help = append([]string{fmt.Sprintf("All commands can be used in private messages without the `%s` prefix.", service.CommandPrefix())}, help...)
				}
			}

			if service.SupportsMultiline() {
				service.SendMessage(message.Channel(), strings.Join(help, "\n"))
			} else {
				for _, h := range help {
					if err := service.SendMessage(message.Channel(), h); err != nil {
						break
					}
				}
			}
		}
	}
}

// Load will load plugin state from a byte array.
func (p *helpPlugin) Load(bot *Bot, service Service, data []byte) error {
	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
		}
	}
	return nil
}

// Save will save plugin state to a byte array.
func (p *helpPlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

// Stats will return the stats for a plugin.
func (p *helpPlugin) Stats(bot *Bot, service Service, message Message) []string {
	return nil
}

// NeHelpPlugin will create a new help plugin.
func NewHelpPlugin() Plugin {
	p := &helpPlugin{}
	return p
}
