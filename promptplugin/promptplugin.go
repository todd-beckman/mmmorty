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

type Prompt struct {
	Prompt string `json: "prompt"`
}

type PromptPlugin struct {
	bot     *mmmorty.Bot
	Prompts []Prompt `json: "prompts"`
}

func (p *PromptPlugin) Help(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message, detailed bool) []string {
	help := mmmorty.CommandHelp(service, addPromptCommand, "some prompt", "adds a prompt for Morty to remember")
	help = append(help, mmmorty.CommandHelp(service, promptCommand, "", "asks Morty for a prompt at random.")[0])
	return help
}

func (p *PromptPlugin) Load(bot *mmmorty.Bot, service mmmorty.Service, data []byte) error {
	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
			return err
		}
	}

	return nil
}

func (p *PromptPlugin) Message(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message) {
	defer bot.MessageRecover(service, message.Channel())

	if service.Name() != mmmorty.DiscordServiceName {
		return
	}

	if service.IsMe(message) {
		return
	}

	if mmmorty.MatchesCommand(service, addPromptCommand, message) {
		p.handleAddPromptCommand(bot, service, message)
	} else if mmmorty.MatchesCommand(service, promptCommand, message) {
		p.handlePromptCommand(bot, service, message)
	}
}

func (p *PromptPlugin) handleAddPromptCommand(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	if service.IsPrivate(message) {
		reply := fmt.Sprintf("Uh, %s, I can't add prompts privately.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	if len(p.Prompts) >= maxPromptCount {
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
		} else {
			promptParts = append(promptParts, word)
		}
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

	p.Prompts = append(p.Prompts, newPrompt)

	reply := fmt.Sprintf("Ok, %s, you got it! I will try to remember that one.", requester)
	service.SendMessage(message.Channel(), reply)
}

func (p *PromptPlugin) handlePromptCommand(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	promptCount := len(p.Prompts)
	if promptCount == 0 {
		reply := fmt.Sprintf("Uh, %s, I don't know any prompts yet. Maybe you could add them?", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	index := rand.Intn(promptCount)
	prompt := p.Prompts[index]
	reply := fmt.Sprintf(promptTemplate, prompt.Prompt)
	service.SendMessage(message.Channel(), reply)
}

func (p *PromptPlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

func (p *PromptPlugin) Stats(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message) []string {
	return []string{}
}

func (p *PromptPlugin) Name() string {
	return "Prompt"
}

func New() mmmorty.Plugin {
	return &PromptPlugin{}
}
