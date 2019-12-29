package wordplugin

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/todd-beckman/mmmorty"
)

const addWordCommand = "add word"
const deleteWordCommand = "forget word"
const defineCommand = "define"

type words struct {
	Words map[string]string `json:"words"`
}

// WordPlugin is the save data for this plugin
type WordPlugin struct {
	bot          *mmmorty.Bot
	WordsByGuild map[string]words `json:"wordsByGuild"`
}

// Help gets the usage for this plugin
func (p *WordPlugin) Help(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage, detailed bool) []string {
	help := mmmorty.CommandHelp(service, defineCommand, "word", "defines the word if I was told to remember it")
	help = append(help, mmmorty.CommandHelp(service, addWordCommand, "word definition", "adds a word I should remember")...)
	help = append(help, mmmorty.CommandHelp(service, deleteWordCommand, "word", "makes me forget a word")...)
	return help
}

// Load loads this plugin from the given data
func (p *WordPlugin) Load(bot *mmmorty.Bot, service mmmorty.Discord, data []byte) error {
	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
			return err
		}
	}

	return nil
}

// Message is the command handler for this plugin
func (p *WordPlugin) Message(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage) {
	defer bot.MessageRecover(service, message.Channel())
	if service.IsMe(message) {
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

	guildID := discordChannel.GuildID
	if p.WordsByGuild == nil {
		p.WordsByGuild = map[string]words{}
	}
	if p.WordsByGuild[guildID].Words == nil {
		p.WordsByGuild[guildID] = words{
			map[string]string{},
		}
	}

	if mmmorty.MatchesCommand(service, addWordCommand, message) {
		p.handleAddWord(bot, service, message, guildID)
	} else if mmmorty.MatchesCommand(service, deleteWordCommand, message) {
		p.handleDeleteWord(bot, service, message, guildID)
	} else if mmmorty.MatchesCommand(service, defineCommand, message) {
		p.handleDefine(bot, service, message, guildID)
	}
}

func (p *WordPlugin) handleAddWord(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage, guildID string) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	if service.IsPrivate(message) {
		reply := fmt.Sprintf("Uh, %s, I can't do this in PM.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	// Should be okay on small servers to let anyone define words.
	// if !service.IsModerator(message) {
	// 	reply := fmt.Sprintf("Uh, %s, I don't think I can let you do that.", requester)
	// 	service.SendMessage(message.Channel(), reply)
	// 	return
	// }

	_, parts := mmmorty.ParseCommand(service, message)

	if len(parts) < 3 {
		reply := fmt.Sprintf("Uh, %s, I need a word and a definition.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	word := strings.ToLower(parts[1])
	definition := strings.Join(parts[2:], " ")

	if old, ok := p.WordsByGuild[guildID].Words[word]; ok {
		reply := fmt.Sprintf("Uh, %s, I added that but overwrote this other one: %q", requester, old)
		service.SendMessage(message.Channel(), reply)
	}
	p.WordsByGuild[guildID].Words[word] = definition

	reply := fmt.Sprintf("You got it, %s! I will try to remember that!", requester)
	service.SendMessage(message.Channel(), reply)
	return
}

func (p *WordPlugin) handleDeleteWord(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage, guildID string) {
	requester := fmt.Sprintf("<@%s>", message.UserID())
	if service.IsPrivate(message) {
		reply := fmt.Sprintf("Uh, %s, I can't do this in PM.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	// Should be okay on small servers to let anyone define words.
	// if !service.IsModerator(message) {
	// 	reply := fmt.Sprintf("Uh, %s, I don't think I can let you do that.", requester)
	// 	service.SendMessage(message.Channel(), reply)
	// 	return
	// }

	_, parts := mmmorty.ParseCommand(service, message)
	if len(parts) != 2 {
		reply := fmt.Sprintf("Uh, %s, I think you forgot to give me word.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	word := strings.ToLower(parts[1])
	if _, ok := p.WordsByGuild[guildID].Words[word]; !ok {
		reply := fmt.Sprintf("Uh, %s, no one told me to remember that word.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	reply := fmt.Sprintf("1... 2... and... poof. I have no idea what %q means.", word)
	service.SendMessage(message.Channel(), reply)
}

func (p *WordPlugin) handleDefine(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage, guildID string) {
	requester := fmt.Sprintf("<@%s>", message.UserID())
	if service.IsPrivate(message) {
		reply := fmt.Sprintf("Uh, %s, I can't do this in PM.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	_, parts := mmmorty.ParseCommand(service, message)
	if len(parts) != 1 {
		reply := fmt.Sprintf("Uh, %s, I think you forgot to name a word.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}
	word := strings.ToLower(parts[0])
	definition, ok := p.WordsByGuild[guildID].Words[word]
	if !ok {
		reply := fmt.Sprintf("Uh, %s, no one told me to remember %s.", requester, word)
		service.SendMessage(message.Channel(), reply)
		return
	}

	reply := fmt.Sprintf("Uh, %s, I think %q is %q.", requester, word, definition)
	service.SendMessage(message.Channel(), reply)
}

// Save will save plugin state to a byte array.
func (p *WordPlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

// Name returns the name of the plugin.
func (p *WordPlugin) Name() string {
	return "Word"
}

// New will create a new Reminder plugin.
func New() mmmorty.Plugin {
	return &WordPlugin{
		WordsByGuild: map[string]words{},
	}
}
