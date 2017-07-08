package colorplugin

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	//"github.com/dustin/go-humanize"
	"github.com/bwmarrin/discordgo"
	"github.com/todd-beckman/mmmorty"
)

const colorCommand = "color me"

type ColorPlugin struct {
	bot *mmmorty.Bot
}

// Used to determine if the role is more than aesthetic
var authPermissions = 0x00000002 | // kick
	0x00000004 | // ban
	0x00000008 | // admin
	0x00000010 | // manage channels
	0x00000020 | // manage guild
	0x00000080 | // view audit log
	0x00002000 | // manage messages
	0x00400000 | // mute members in voice channel
	0x00800000 | // deafen members in voice channel
	0x01000000 | // move members between channels
	0x08000000 | // modify others' nicknames
	0x10000000 | // manage roles
	0x20000000 | // manage webhooks
	0x40000000 //   manage emojis

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

	if service.Name() != mmmorty.DiscordServiceName {
		return
	}

	if service.IsMe(message) {
		return
	}

	if !mmmorty.MatchesCommand(service, "color me", message) {
		return
	}

	_, parts := mmmorty.ParseCommand(service, message)

	requester := message.UserName()
	if service.Name() == mmmorty.DiscordServiceName {
		requester = fmt.Sprintf("<@%s>", message.UserID())
	}

	if len(parts) != 2 {
		reply := fmt.Sprintf("Uh, %s, I can't give you more than one color.")
		service.SendMessage(message.Channel(), reply)
		return
	}

	color := strings.ToLower(parts[1])

	discord := service.(*mmmorty.Discord)
	discordMessage := message.(mmmorty.DiscordMessage)
	roles := discord.GetRoles(discordMessage)
	var role *discordgo.Role
	for _, r := range roles {
		if strings.ToLower(r.Name) == color {
			role = r
			break
		}
	}

	if role == nil {
		reply := fmt.Sprintf("Uh, Rick, I can't find a role called %s", color)
		service.SendMessage(message.Channel(), reply)
		return
	}

	if role.Permissions&authPermissions > 0 {
		reply := fmt.Sprintf("Uh, Rick, I think  %s is more than just a colored role.")
		service.SendMessage(message.Channel(), reply)
		return
	}

	reply := fmt.Sprintf("Sorry %s, you want to be %s but I can't do that yet", requester, color)
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
