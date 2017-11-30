package diceplugin

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	"github.com/todd-beckman/mmmorty"
)

const (
	maxWordCount       = 100
	rollCommand        = "roll"
	simpleRollTemplate = "Uh, %s, it looks like it landed on %s"

	shorthandRollTemplate = "Uh, %s, it looks like they landed on %s which makes %s"
)

var (
	simpleRollRegex    = regexp.MustCompile(`\d+`)
	shorthandRollRegex = regexp.MustCompile(`\d*d\d+`)
)

type DicePlugin struct {
	bot *mmmorty.Bot
}

func (p *DicePlugin) Help(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message, detailed bool) []string {
	return mmmorty.CommandHelp(service, rollCommand, "X sided die OR roll XdY",
		"asks Morty to roll dice for you")
}

func (p *DicePlugin) Load(bot *mmmorty.Bot, service mmmorty.Service, data []byte) error {
	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
			return err
		}
	}

	return nil
}

func (p *DicePlugin) Message(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message) {
	defer bot.MessageRecover(service, message.Channel())

	if service.Name() != mmmorty.DiscordServiceName {
		return
	}

	if service.IsMe(message) {
		return
	}

	if mmmorty.MatchesCommand(service, rollCommand, message) {
		p.handleRollCommand(bot, service, message)
	}
}

func (p *DicePlugin) handleRollCommand(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	_, parts := mmmorty.ParseCommand(service, message)

	if len(parts) < 0 {
		reply := fmt.Sprintf("Uh, %s, could you tell me what to roll? `roll X sided die` or `roll XdY` should work.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	param := parts[0]
	if ok := shorthandRollRegex.MatchString(param); ok {
		p.handleShorthandRollCommand(bot, service, message, parts)
		return
	}

	if ok := simpleRollRegex.MatchString(param); ok {
		p.handleSimpleRollCommand(bot, service, message, parts)
		return
	}

	reply := fmt.Sprintf("Uh, %s, I don't get that. Try `roll X sided die` or `roll XdY` should work.", requester)
	service.SendMessage(message.Channel(), reply)
}

func (p *DicePlugin) handleSimpleRollCommand(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message, parts []string) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	sides, err := strconv.Atoi(parts[0])
	if err != nil || sides < 1 {
		reply := fmt.Sprintf("U1h, %s, I don't think I can roll a die with %s sides.", requester, parts[0])
		service.SendMessage(message.Channel(), reply)
		return
	}

	roll := strconv.Itoa(rand.Intn(sides) + 1)

	reply := fmt.Sprintf(simpleRollTemplate, requester, roll)
	service.SendMessage(message.Channel(), reply)
}

func (p *DicePlugin) handleShorthandRollCommand(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message, parts []string) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	shorthand := strings.Split(parts[0], "d")

	var dice int
	if shorthand[0] == "" {
		dice = 1
	} else {
		dice, _ = strconv.Atoi(shorthand[0])
	}
	sides, _ := strconv.Atoi(shorthand[1])

	if dice < 1 {
		reply := fmt.Sprintf("Uh, %s, I don't think I can roll %s dice.", requester, shorthand[0])
		service.SendMessage(message.Channel(), reply)
		return
	}

	if sides < 1 {
		reply := fmt.Sprintf("Uh, %s, I don't think I can roll a die with %s sides.", requester, shorthand[1])
		service.SendMessage(message.Channel(), reply)
		return
	}

	rolls := []string{}
	sum := 0
	for i := 0; i < dice; i++ {
		roll := rand.Intn(sides) + 1
		rolls = append(rolls, strconv.Itoa(roll))
		sum += roll
	}

	results := strings.Join(rolls, " + ")

	reply := fmt.Sprintf(shorthandRollTemplate, requester, results, strconv.Itoa(sum))
	service.SendMessage(message.Channel(), reply)
}

func (p *DicePlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

func (p *DicePlugin) Stats(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message) []string {
	return []string{}
}

func (p *DicePlugin) Name() string {
	return "Dice"
}

func New() mmmorty.Plugin {
	return &DicePlugin{}
}
