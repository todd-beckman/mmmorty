package promptplugin

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/todd-beckman/mmmorty"
)

const (
	addPromptCommand = "add prompt"
	promptCommand    = "prompt"
	promptTemplate   = "```\n" +
		"%s\n" +
		"```"

	maxPromptCount = 500
	maxWordCount   = 150
)

// Prompt is a prompt
type Prompt struct {
	Prompt string `json:"prompt"`
}

// PromptPlugin is the save structure of this plugin
type PromptPlugin struct {
	bot     *mmmorty.Bot
	Prompts map[string][]Prompt `json:"prompts"`
}

// Help gets the usage for this plugin
func (p *PromptPlugin) Help(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage, detailed bool) []string {
	help := mmmorty.CommandHelp(service, addPromptCommand, "some prompt", "adds a prompt for Morty to remember")
	help = append(help, mmmorty.CommandHelp(service, promptCommand, "", "asks Morty for a prompt at random.")[0])
	return help
}

// Load sets the state of the plugin from the given data
func (p *PromptPlugin) Load(bot *mmmorty.Bot, service mmmorty.Discord, data []byte) error {
	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
			return err
		}
	}

	return nil
}

// Message is the command handler for this plugin
func (p *PromptPlugin) Message(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage) {
	defer bot.MessageRecover(service, message.Channel())

	if service.IsMe(message) {
		return
	}

	channelID := message.Channel()
	discordChannel, err := service.Channel(channelID)
	if err != nil {
		requester := fmt.Sprintf("<@%s>", message.UserID())
		reply := fmt.Sprintf("Uh, %s, something went figuring out your server.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}
	guildID := discordChannel.GuildID

	if mmmorty.MatchesCommand(service, addPromptCommand, message) {
		p.handleAddPromptCommand(bot, service, message, guildID)
	} else if mmmorty.MatchesCommand(service, promptCommand, message) {
		p.handlePromptCommand(bot, service, message, guildID)
	}
}

func (p *PromptPlugin) handleAddPromptCommand(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage, guildID string) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	if len(p.Prompts[guildID]) >= maxPromptCount {
		reply := fmt.Sprintf("Uh, %s, I can't remember all these prompts. Rick might need to help get rid of some.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	if strings.Contains(message.Message(), "http") {
		reply := fmt.Sprintf("Uh, %s, I would rather not remember prompts with links.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	_, parts := mmmorty.ParseCommand(service, message)

	promptParts := []string{}
	for index, word := range parts[1:] {
		if index > maxWordCount {
			reply := fmt.Sprintf("Uh, %s, that prompt is kind of long. Is there any way you can shorten it?", requester)
			service.SendMessage(message.Channel(), reply)
			return
		}
		promptParts = append(promptParts, word)
	}

	if len(promptParts) == 0 {
		reply := fmt.Sprintf("Uh, %s, I didn't get that. Be sure to say the prompt you want me to remember.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	plotPrompt := strings.Join(promptParts, " ")

	newPrompt := Prompt{
		Prompt: plotPrompt,
	}

	if p.Prompts == nil {
		p.Prompts = map[string][]Prompt{
			guildID: []Prompt{},
		}
	}
	if p.Prompts[guildID] == nil {
		p.Prompts[guildID] = []Prompt{}
	}

	p.Prompts[guildID] = append(p.Prompts[guildID], newPrompt)

	reply := fmt.Sprintf("Ok, %s, you got it! I will try to remember that one.", requester)
	service.SendMessage(message.Channel(), reply)
}

func (p *PromptPlugin) handlePromptCommand(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage, guildID string) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	promptCount := len(p.Prompts[guildID])
	if promptCount == 0 {
		reply := fmt.Sprintf("Uh, %s, I don't know any prompts yet. Maybe you could add them?", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	index := rand.Intn(promptCount)
	prompt := p.Prompts[guildID][index]
	reply := fmt.Sprintf(promptTemplate, prompt.Prompt)
	service.SendMessage(message.Channel(), reply)
}

// Save saves this plugin
func (p *PromptPlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

// Name gets the name of the plugin for saving purposes
func (p *PromptPlugin) Name() string {
	return "Prompt"
}

// New creates a new instance of this plugin.
func New() mmmorty.Plugin {
	return &PromptPlugin{}
}
