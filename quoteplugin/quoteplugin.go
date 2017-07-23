package quoteplugin

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/todd-beckman/mmmorty"
)

const (
	addQuoteCommand = "add quote"
	quoteCommand    = "quote me"
	quoteTemplate   = "```\n" +
		"%s said:\n" +
		"%s\n" +
		"```"

	maxQuoteCount = 500
	maxWordCount  = 150
)

type Quote struct {
	Author string `json: "author"`
	Quote  string `json: "quote"`
}

type QuotePlugin struct {
	bot    *mmmorty.Bot
	Quotes []Quote `json: "quotes"`
}

func (p *QuotePlugin) Help(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message, detailed bool) []string {
	help := mmmorty.CommandHelp(service, addQuoteCommand, "<author> said <quote>", "adds a quote for Morty to remember")
	help = append(help, mmmorty.CommandHelp(service, quoteCommand, "", "retrieves a quote at random.")[0])
	return help
}

func (p *QuotePlugin) Load(bot *mmmorty.Bot, service mmmorty.Service, data []byte) error {
	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
			return err
		}
	}

	return nil
}

func (p *QuotePlugin) Message(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message) {
	defer bot.MessageRecover(service, message.Channel())

	if service.Name() != mmmorty.DiscordServiceName {
		return
	}

	if service.IsMe(message) {
		return
	}

	if mmmorty.MatchesCommand(service, addQuoteCommand, message) {
		p.handleAddQuoteCommand(bot, service, message)
	} else if mmmorty.MatchesCommand(service, quoteCommand, message) {
		p.handleQuoteCommand(bot, service, message)
	}
}

func (p *QuotePlugin) handleAddQuoteCommand(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	if service.IsPrivate(message) {
		reply := fmt.Sprintf("Uh, %s, I can't add quotes privately.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	if len(p.Quotes) >= maxQuoteCount {
		reply := fmt.Sprintf("Uh, %s, I can't remember all these quotes. Rick might need to help get rid of some.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	if strings.Contains(message.Message(), "http") {
		reply := fmt.Sprintf("Uh, %s, I would rather not remember quotes with links.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	_, parts := mmmorty.ParseCommand(service, message)

	authorParts := []string{}
	quoteParts := []string{}
	saidIndex := 0
	for index, word := range parts[1:] { // first word is 'quote' because 'add quote'
		if index > maxWordCount {
			reply := fmt.Sprintf("Uh, %s, that quote is kind of long. Is there any way you can shorten it?", requester)
			service.SendMessage(message.Channel(), reply)
			return
		}
		if saidIndex == 0 && word == "said" {
			saidIndex = index
		} else if saidIndex == 0 {
			authorParts = append(authorParts, word)
		} else {
			quoteParts = append(quoteParts, word)
		}
	}

	if len(quoteParts) == 0 {
		reply := fmt.Sprintf("Uh, %s, I didn't get that. Maybe put `said` between the author and the quote?", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	author := strings.Join(authorParts, " ")
	quote := strings.Join(quoteParts, " ")

	newQuote := Quote{
		Author: author,
		Quote:  quote,
	}

	p.Quotes = append(p.Quotes, newQuote)

	reply := fmt.Sprintf("Ok, %s, you got it! I will try to remember that one.", requester)
	service.SendMessage(message.Channel(), reply)
}

func (p *QuotePlugin) handleQuoteCommand(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	quoteCount := len(p.Quotes)
	if quoteCount == 0 {
		reply := fmt.Sprintf("Uh, %s, I don't know any quotes yet. Maybe you could add them?", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	index := rand.Intn(quoteCount)
	quote := p.Quotes[index]
	reply := fmt.Sprintf(quoteTemplate, quote.Author, quote.Quote)
	service.SendMessage(message.Channel(), reply)
}

func (p *QuotePlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

func (p *QuotePlugin) Stats(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message) []string {
	return []string{}
}

func (p *QuotePlugin) Name() string {
	return "Quote"
}

func New() mmmorty.Plugin {
	return &QuotePlugin{}
}
