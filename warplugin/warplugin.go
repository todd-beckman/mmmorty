package quoteplugin

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/todd-beckman/mmmorty"
)

const (
	startWarCommand = "start sprint"
	endWarCommand   = "end sprint"
	joinWarCommand  = "join sprint"
	leaveWarCommand = "leave sprint"
	maxWarCount     = 10
)

type War struct {
	// public
	Channel   string   `json: "channel"`  // which channel the sprint was started from
	Duration  int      `json: "duration"` // minutes
	Name      string   `json: "name"`
	Sprinters []string `json: "sprinters"` // list of user ID's of players to ping for updates
	Start     int64    `json: "start"`     // Unix time, the number of seconds elapsed since January 1, 1970 UTC.

	// private
	alertTimer *time.Timer
	startTimer *time.Timer
	endTimer   *time.Timer
}

type WarPlugin struct {
	bot  *mmmorty.Bot
	Wars map[string]*War `json: "wars"` // map of name to war
}

func timeWithoutSeconds() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location)
}

func stringifySprinters(war *War) string {
	notifyUsers := []string{}
	for _, sprinter := range war.Sprinters {
		notifyUsers = append(notifyUsers, fmt.Sprintf("<@%s>", sprinter))
	}
	return strings.Join(notifyUsers, " ")
}

func (p *WarPlugin) pickName() string {
	var name string
	for p.Wars[name] != nil {
		name = fmt.Sprintf(
			"%s%s%s",
			rand.Intn(10),
			rand.Intn(10),
			rand.Intn(10),
		)
	}

	return name
}

func (p *WarPlugin) alertNotify(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message, name string) {
	war, ok := p.Wars[war]
	if !ok {
		return
	}

	notifyString := stringifySprinters(war)
	reply := fmt.Sprintf("Sprint %s is starting in one minute! %s", name, notifyString)
	service.SendMessage(war.Channel)
}

func (p *WarPlugin) startNotify(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message, name string) {
	war, ok := p.Wars[war]
	if !ok {
		return
	}

	notifyString := stringifySprinters(war)
	reply := fmt.Sprintf("Sprint %s starts now! %s", name, notifyString)
	service.SendMessage(war.Channel)
}

func (p *WarPlugin) endNotify(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message, name string) {
	war, ok := p.Wars[war]
	if !ok {
		return
	}

	notifyString := stringifySprinters(war)
	reply := fmt.Sprintf("Sprint %s has ended! %s", name, notifyString)
	service.SendMessage(war.Channel)

	delete(p.Wars, name)
}

func (p *WarPlugin) Help(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message, detailed bool) []string {
	help := mmmorty.CommandHelp(service, startWarCommand,
		"at :XX for Y (mins)", "starts a sprint starting when the minute hand points to XX and lasting for Y minutes")
	help = append(help, mmmorty.CommandHelp(service, endWarCommand, "<id>",
		"Ends the sprint with the given name.")[0])
	help = append(help, mmmorty.CommandHelp(service, joinWarCommand, "<id>",
		"Adds you to the list of people to notify for the given sprint.")[0])
	help = append(help, mmmorty.CommandHelp(service, leaveWarCommand, "<id>",
		"Removes you from the list of people to notify for the given sprint.")[0])
	return help
}

func (p *WarPlugin) Load(bot *mmmorty.Bot, service mmmorty.Service, data []byte) error {
	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
			return err
		}
	}

	return nil
}

func (p *WarPlugin) Message(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message) {
	defer mmmorty.MessageRecover()

	if service.Name() != mmmorty.DiscordServiceName {
		return
	}

	if service.IsMe(message) {
		return
	}

	if mmmorty.MatchesCommand(service, startWarCommand, message) {
		p.handleStartWarCommand(bot, service, message)
	}
}

func (p *WarPlugin) handleStartWarCommand(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	if service.IsPrivate(message) {
		reply := fmt.Sprintf("Uh, %s, I can't start sprints privately.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	if len(p.Quotes) >= maxQuoteCount {
		reply := fmt.Sprintf("Uh, %s, I can't remember all these sprints. Could you end some first?", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}
	_, parts := mmmorty.ParseCommand(service, message)

	// "sprint at :XX for Y (mins)"
	if len(parts) < 5 || parts[1] != "at" || parts[3] != "for" {
		reply := fmt.Sprintf("Uh, %s, I don't quite know what you mean. Have you tried `%s at :XX for Y (mins)`?", requester, startWarCommand)
		service.SendMessage(message.Channel(), reply)
		return
	}

	// get the :XX minutes portion
	requestedTime := parts[2]
	if requestedTime[0] == ':' {
		requestedTime = requestedTime[1:]
	}
	minutes, err := strconv.Atoi(requestedTime)
	if err != nil {
		reply := fmt.Sprintf("Uh, %s, that start time doesn't make sense. `:XX` should work fine.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}
	if minutes < 0 || minutes > 60 {
		reply := fmt.Sprintf("Uh, %s, that start time doesn't make sense. Maybe try a number between 0 and 60, you know?", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	// get the Y (mins) portion
	requestedDuration := parts[4]
	duration, err := strconv.Atoi(requestedDuration)
	if err != nil {
		reply := fmt.Sprintf("Uh, %s, that duration doesn't make sense. A number for the minutes should work fine.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}
	if duration > 180 {
		reply := fmt.Sprintf("Gee, %s, I don't know if Rick will let me keep a sprint going for that long.", requester)
		service.SendMessage(message.Channel(), reply)
	}

	now := timeWithoutSeconds()
	nowMinutes := now.Minute()

	// Cannot start the timer for the current minute
	if minutes == nowMinutes {
		// TODO: print error
		return
	}

	// Rollover to the next hour
	if minutes < nowMinutes {
		minutes += 60
	}

	// unique ID used to remember this war
	name := p.pickName()

	// when to give the minute-before alert
	minutesToAlert := minutes - nowMinutes - 1
	var alertTimer *time.Timer
	if minutesToAlert > 0 {
		alertingIn := time.ParseDuration(fmt.Sprintf("%sm", minutesToAlert))
		alertTimer = time.AfterFunc(alertingIn, func() {
			p.alertNotify(bot, service, message, name)
		})
	}

	// when to give the starting alert
	minutesToStart := minutes - nowMinutes
	startingIn := time.ParseDuration(fmt.Sprintf("%sm", minutesToStart))
	startTimer := time.AfterFunc(startingIn, func() {
		p.startNotify(bot, service, message, name)
	})

	// when to give the ending alert
	minutesToEnd := minutesFromNow + duration
	endingIn := time.ParseDuration(fmt.Sprintf("%sm", minutesToEnd))
	endTimer := time.AfterFunc(endingIn, func() {
		p.endNotify(bot, service, message, name)
	})

	// backup the war info
	war := &War{
		Channel:   message.Channel(),
		Duration:  duration,
		Name:      name,
		Sprinters: []string{message.UserID()},
		Start:     startingTime.Unix(),

		alertTimer: alertTimer,
		startTimer: startTimer,
		endTimer:   endTimer,
	}
	p.Wars[name] = war

	reply := fmt.Sprintf("Ok, %s, you got it! I added you to this sprint. Use `%s %s` to get updates, `%s %s` to stop getting them, and `%s %s` to cancel this sprint.")
	service.SendMessage(message.Channel(), reply)
}

func (p *WarPlugin) handleEndWarCommand(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	_, parts := mmmorty.ParseCommand(service, message)
	if len(parts) < 2 {
		reply := fmt.Sprintf("Uh, %s, what was the sprint you wanted to end?", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	name := parts[1]

	war, ok := p.Wars[name]
	if !ok {
		reply := fmt.Sprintf("Uh, %s, I don't see a war by that name.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

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

func (p *WarPlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

func (p *WarPlugin) Stats(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message) []string {
	return []string{}
}

func (p *WarPlugin) Name() string {
	return "Quote"
}

func New() mmmorty.Plugin {
	return &WarPlugin{}
}
