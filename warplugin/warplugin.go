package warplugin

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
	endWarCommand   = "end"
	joinWarCommand  = "join"
	leaveWarCommand = "leave"
	doTheThing      = "do the thing"
	maxWarCount     = 10
)

// War is a timed sprint with users subscribed to the start and end alerts
type War struct {
	// public
	Channel   string   `json:"channel"`  // which channel the sprint was started from
	Duration  int      `json:"duration"` // minutes
	Name      string   `json:"name"`
	Sprinters []string `json:"sprinters"` // list of user ID's of players to ping for updates
	Start     int64    `json:"start"`     // Unix time, the number of seconds elapsed since January 1, 1970 UTC.

	// private
	alertTimer *time.Timer
	startTimer *time.Timer
	endTimer   *time.Timer
}

// WarPlugin is this plugin's save structure
type WarPlugin struct {
	bot  *mmmorty.Bot
	Wars map[string]*War `json:"wars"` // map of name to war
}

// Help gets the usage info for this plugin
func (p *WarPlugin) Help(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage, detailed bool) []string {
	help := mmmorty.CommandHelp(
		service, startWarCommand, "at :XX for Y (mins)",
		"starts a sprint starting when the minute hand points to XX and lasting for Y minutes",
	)
	help = append(help, mmmorty.CommandHelp(
		service, endWarCommand, "ID",
		"Ends the sprint with the given name.",
	)[0])
	help = append(help, mmmorty.CommandHelp(
		service, joinWarCommand, "ID",
		"Adds you to the list of people to notify for the given sprint.",
	)[0])
	help = append(help, mmmorty.CommandHelp(
		service, leaveWarCommand, "ID",
		"Removes you from the list of people to notify for the given sprint.",
	)[0])
	help = append(help, mmmorty.CommandHelp(
		service, doTheThing, "",
		"Shorthand for \"start sprint for 15\" starting in 4 minutes.",
	)[0])
	return help
}

// Load loads the plugin with the given data
func (p *WarPlugin) Load(bot *mmmorty.Bot, service mmmorty.Discord, data []byte) error {
	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
			return err
		}
	}

	return nil
}

// Message is the command handler for this plugin
func (p *WarPlugin) Message(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage) {
	defer bot.MessageRecover(service, message.Channel())

	if service.Name() != mmmorty.DiscordServiceName {
		return
	}

	if service.IsMe(message) {
		return
	}

	if mmmorty.MatchesCommand(service, startWarCommand, message) {
		p.handleStartWarCommand(bot, service, message)
	} else if mmmorty.MatchesCommand(service, doTheThing, message) {
		p.handleDoTheThing(bot, service, message)
	} else if mmmorty.MatchesCommand(service, joinWarCommand, message) {
		p.handleJoinWarCommand(bot, service, message)
	} else if mmmorty.MatchesCommand(service, leaveWarCommand, message) {
		p.handleLeaveWarCommand(bot, service, message)
	} else if mmmorty.MatchesCommand(service, endWarCommand, message) {
		p.handleEndWarCommand(bot, service, message)
	}
}

func (p *WarPlugin) handleDoTheThing(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage) {
	now := timeWithoutSeconds()
	nowMinute := now.Minute()

	startMinute := (nowMinute + 4) % 60

	p.startWar(bot, service, message, startMinute, 15)
}

func (p *WarPlugin) handleStartWarCommand(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	if service.IsPrivate(message) {
		reply := fmt.Sprintf("Uh, %s, I can't start sprints privately.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	_, parts := mmmorty.ParseCommand(service, message)

	// "sprint at :XX for Y (mins)"
	if len(parts) < 5 || (parts[1] != "at" && parts[1] != "for") || (parts[3] != "at" && parts[3] != "for") {
		reply := fmt.Sprintf("Uh, %s, I don't quite know what you mean. Have you tried `%s at :XX for Y (mins)`?", requester, startWarCommand)
		service.SendMessage(message.Channel(), reply)
		return
	}

	var requestedTime, requestedDuration string
	if parts[1] == "at" && parts[3] == "for" {
		requestedTime = parts[2]
		requestedDuration = parts[4]
	} else if parts[1] == "for" && parts[3] == "at" {
		requestedDuration = parts[2]
		requestedTime = parts[4]
	} else {
		reply := fmt.Sprintf("Uh, %s, I don't quite know what you mean. Have you tried `%s at :XX for Y (mins)`?", requester, startWarCommand)
		service.SendMessage(message.Channel(), reply)
		return
	}

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

	duration, err := strconv.Atoi(requestedDuration)
	if err != nil {
		reply := fmt.Sprintf("Uh, %s, that duration doesn't make sense. A number for the minutes should work fine.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}
	if duration > 180 {
		reply := fmt.Sprintf("Gee, %s, I don't know if Rick will let me keep a sprint going for that long.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	p.startWar(bot, service, message, minutes, duration)
}

func (p *WarPlugin) getNameFromParts(parts []string) string {
	if len(parts) < 1 {
		if len(p.Wars) != 1 {
			return ""
		}

		// Guaranteed length  1
		for k := range p.Wars {
			return k
		}
	}
	return parts[0]
}

func (p *WarPlugin) handleJoinWarCommand(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	_, parts := mmmorty.ParseCommand(service, message)
	name := p.getNameFromParts(parts)

	if name == "" {
		reply := fmt.Sprintf("Uh, %s, what was the sprint you wanted to join?", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	war, ok := p.Wars[name]
	if !ok {
		reply := fmt.Sprintf("Uh, %s, I don't see a sprint by that name.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	isInWar := false
	for _, sprinter := range war.Sprinters {
		if sprinter == message.UserID() {
			isInWar = true
			break
		}
	}

	if isInWar {
		reply := fmt.Sprintf("Looks like you are in that sprint already %s. You should be good to go.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	war.Sprinters = append(war.Sprinters, message.UserID())

	reply := fmt.Sprintf("I added you to the sprint, %s. Good luck!", requester)
	service.SendMessage(message.Channel(), reply)
}

func (p *WarPlugin) handleLeaveWarCommand(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	_, parts := mmmorty.ParseCommand(service, message)
	name := p.getNameFromParts(parts)

	if name == "" {
		reply := fmt.Sprintf("Uh, %s, what was the sprint you wanted to leave?", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	war, ok := p.Wars[name]
	if !ok {
		reply := fmt.Sprintf("Uh, %s, I don't see a sprint by that name.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	index := -1
	for i, sprinter := range war.Sprinters {
		if sprinter == message.UserID() {
			index = i
			break
		}
	}

	if index == -1 {
		reply := fmt.Sprintf("Uh, %s, you are not in that sprint.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	if len(war.Sprinters)-1 == index {
		war.Sprinters = war.Sprinters[:index]
	} else if index == 0 {
		war.Sprinters = war.Sprinters[index+1:]
	} else {
		war.Sprinters = append(war.Sprinters[:index], war.Sprinters[index+1:]...)
	}

	reply := fmt.Sprintf("I removed you from %s.", name)
	service.SendMessage(message.Channel(), reply)
}

func (p *WarPlugin) handleEndWarCommand(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	_, parts := mmmorty.ParseCommand(service, message)
	name := p.getNameFromParts(parts)

	if name == "" {
		reply := fmt.Sprintf("Uh, %s, what was the sprint you wanted to end?", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	war, ok := p.Wars[name]
	if !ok {
		reply := fmt.Sprintf("Uh, %s, I don't see a sprint by that name.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	war.alertTimer.Stop()
	war.startTimer.Stop()
	war.endTimer.Stop()
	delete(p.Wars, name)

	reply := fmt.Sprintf("Sprint %s was ended.", name)
	service.SendMessage(message.Channel(), reply)
}

// Save saves the state of the plugin to file
func (p *WarPlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

// Name gets the name of this plugin for saving purposes
func (p *WarPlugin) Name() string {
	return "War"
}

// New creates a new instance of this plugin
func New() mmmorty.Plugin {
	return &WarPlugin{
		Wars: map[string]*War{},
	}
}

func (p *WarPlugin) startWar(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage, minutes, duration int) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

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
		alertingIn, _ := time.ParseDuration(fmt.Sprintf("%vm", minutesToAlert))
		alertTimer = time.AfterFunc(alertingIn, func() {
			p.alertNotify(bot, service, message, name)
		})
	}

	// when to give the starting alert
	minutesToStart := minutes - nowMinutes
	startingIn, _ := time.ParseDuration(fmt.Sprintf("%vm", minutesToStart))
	startTimer := time.AfterFunc(startingIn, func() {
		p.startNotify(bot, service, message, name)
	})

	// when to give the ending alert
	minutesToEnd := minutesToStart + duration
	endingIn, _ := time.ParseDuration(fmt.Sprintf("%vm", minutesToEnd))
	endTimer := time.AfterFunc(endingIn, func() {
		p.endNotify(bot, service, message, name)
	})

	// backup the war info
	war := &War{
		Channel:   message.Channel(),
		Duration:  duration,
		Name:      name,
		Sprinters: []string{message.UserID()},
		Start:     now.Add(startingIn).Unix(),

		alertTimer: alertTimer,
		startTimer: startTimer,
		endTimer:   endTimer,
	}
	p.Wars[name] = war

	reply := fmt.Sprintf(
		"Ok, %s, you got it! I added you to this sprint. Use `%s %s` to get updates, `%s %s` to stop getting them, and `%s %s` to cancel this sprint.",
		requester,
		joinWarCommand, name,
		leaveWarCommand, name,
		endWarCommand, name,
	)
	service.SendMessage(message.Channel(), reply)
}

func timeWithoutSeconds() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())
}

func stringifySprinters(war *War) string {
	notifyUsers := []string{}
	for _, sprinter := range war.Sprinters {
		notifyUsers = append(notifyUsers, fmt.Sprintf("<@%s>", sprinter))
	}
	return strings.Join(notifyUsers, " ")
}

func (p *WarPlugin) pickName() string {
	name := fmt.Sprintf(
		"%v%v%v",
		rand.Intn(10),
		rand.Intn(10),
		rand.Intn(10),
	)
	_, ok := p.Wars[name]
	if ok {
		return p.pickName()
	}

	return name
}

func (p *WarPlugin) alertNotify(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage, name string) {
	war, ok := p.Wars[name]
	if !ok {
		return
	}

	notifyString := stringifySprinters(war)

	duration := war.Duration
	minuteString := "minutes"
	if duration == 1 {
		minuteString = "minute"
	}

	reply := fmt.Sprintf("Sprint %s is starting in one minute, when it will go for %v %s! %s", name, duration, minuteString, notifyString)
	service.SendMessage(war.Channel, reply)
}

func (p *WarPlugin) startNotify(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage, name string) {
	war, ok := p.Wars[name]
	if !ok {
		return
	}

	notifyString := stringifySprinters(war)

	duration := war.Duration
	minuteString := "minutes"
	if duration == 1 {
		minuteString = "minute"
	}
	reply := fmt.Sprintf("Sprint %s starts now and goes for %v %s! %s", name, duration, minuteString, notifyString)
	service.SendMessage(war.Channel, reply)
}

func (p *WarPlugin) endNotify(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage, name string) {
	war, ok := p.Wars[name]
	if !ok {
		return
	}

	notifyString := stringifySprinters(war)
	reply := fmt.Sprintf("Sprint %s has ended! %s", name, notifyString)
	service.SendMessage(war.Channel, reply)

	delete(p.Wars, name)
}
