package colorplugin

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/todd-beckman/mmmorty"
)

const colorCommand = "color me"
const manageColorCommand = "manage color"
const stopManagingCommand = "stop managing"

type colorSet struct {
	ManagedRoles map[string]bool `json:"managedRoles"`
}

// ColorPlugin is the save data for this plugin
type ColorPlugin struct {
	bot          *mmmorty.Bot
	RolesByGuild map[string]colorSet `json:"rolesByGuild"`
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

func doesRoleHaveAuth(permissions int) bool {
	return permissions&authPermissions > 0
}

func (p *ColorPlugin) getPrintableRoles(guildID string) []string {
	printableRoles := []string{}
	for role, isManaged := range p.RolesByGuild[guildID].ManagedRoles {
		if isManaged {
			printableRoles = append(printableRoles, role)
		}
	}
	return printableRoles
}

// Help gets the usage for this plugin
func (p *ColorPlugin) Help(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage, detailed bool) []string {
	help := mmmorty.CommandHelp(service, colorCommand, "color", "assigns the desired color if this server supports it and the color is available")
	return help
}

// Load loads this plugin from the given data
func (p *ColorPlugin) Load(bot *mmmorty.Bot, service mmmorty.Discord, data []byte) error {
	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
			return err
		}
	}

	return nil
}

// Message is the command handler for this plugin
func (p *ColorPlugin) Message(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage) {
	defer bot.MessageRecover(service, message.Channel())

	if service.IsMe(message) {
		return
	}

	requester := fmt.Sprintf("<@%s>", message.UserID())

	if service.IsMe(message) {
		return
	}

	channelID := message.Channel()
	discordChannel, err := service.Channel(channelID)
	if err != nil {
		reply := fmt.Sprintf("Uh, %s, something went figuring out your server.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}
	guildID := discordChannel.GuildID

	if p.RolesByGuild == nil {
		p.RolesByGuild = map[string]colorSet{
			guildID: colorSet{},
		}
	}

	if p.RolesByGuild[guildID].ManagedRoles == nil {
		p.RolesByGuild[guildID] = colorSet{
			ManagedRoles: map[string]bool{},
		}
	}

	if mmmorty.MatchesCommand(service, colorCommand, message) {
		p.handleColorMe(bot, service, message, guildID)
	} else if mmmorty.MatchesCommand(service, manageColorCommand, message) {
		p.handleManageColor(bot, service, message, guildID)
	} else if mmmorty.MatchesCommand(service, stopManagingCommand, message) {
		p.handleStopManaging(bot, service, message, guildID)
	}
}

func (p *ColorPlugin) handleColorMe(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage, guildID string) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	if service.IsPrivate(message) {
		reply := fmt.Sprintf("Uh, %s, I cannot color you in private.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	if availableRoles := p.getPrintableRoles(guildID); len(availableRoles) == 0 {
		reply := fmt.Sprintf("Uh, %s, I don't think this server lets me set your color.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	_, parts := mmmorty.ParseCommand(service, message)

	if len(parts) == 1 {
		reply := fmt.Sprintf("Uh, %s, I think you forgot to name a color.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	} else if len(parts) > 2 {
		reply := fmt.Sprintf("Uh, %s, I can't give you more than one color.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	color := strings.ToLower(parts[1])

	role := service.GetRoleByName(message.Channel(), color)

	if role == nil {
		reply := fmt.Sprintf("Uh, %s, I can't find a role called %s", requester, color)
		service.SendMessage(message.Channel(), reply)
		return
	}

	if doesRoleHaveAuth(role.Permissions) {
		reply := fmt.Sprintf("Uh, %s, I think %s is more than just a colored role.", requester, color)
		service.SendMessage(message.Channel(), reply)
		return
	}

	// Remove all managed roles first so the user doesn't have multiple colors
	userRoles := service.UserRoles(guildID, message.UserID())
	for _, userRole := range userRoles {
		for r, isManaged := range p.RolesByGuild[guildID].ManagedRoles {
			if !isManaged {
				continue
			}

			managedRole := service.GetRoleByName(message.Channel(), r)
			if userRole == managedRole.ID {
				ok := service.GuildMemberRoleRemove(guildID, message.UserID(), userRole)
				if !ok {
					reply := fmt.Sprintf("Uh, %s, something went wrong. Are you sure I can manage %v?", requester, color)
					service.SendMessage(message.Channel(), reply)
					continue
				}
			}
		}
	}

	ok := service.GuildMemberRoleAdd(guildID, message.UserID(), role.ID)
	if !ok {
		reply := fmt.Sprintf("Uh, %s, something went wrong. Are you sure I can let you be %v?", requester, color)
		service.SendMessage(message.Channel(), reply)
		return
	}

	reply := fmt.Sprintf("You got it, %s! You are now %s", requester, color)
	service.SendMessage(message.Channel(), reply)
	return

}

func (p *ColorPlugin) handleManageColor(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage, guildID string) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	if service.IsPrivate(message) {
		reply := fmt.Sprintf("Uh, %s, I cannot color you in private.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	if message.UserID() != service.OwnerUserID {
		reply := fmt.Sprintf("Uh, %s, I think you need to ask my Rick for that command.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	_, parts := mmmorty.ParseCommand(service, message)

	if len(parts) == 1 {
		reply := fmt.Sprintf("Uh, %s, I think you forgot to name a color.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	for _, c := range parts[1:] {
		color := strings.ToLower(c)
		if p.RolesByGuild[guildID].ManagedRoles[color] {
			reply := fmt.Sprintf("Uh, %s, I am already managing %s", requester, color)
			service.SendMessage(message.Channel(), reply)
			continue
		}

		role := service.GetRoleByName(message.Channel(), color)

		if role == nil {
			reply := fmt.Sprintf("Uh, %s, I can't find a role called %s", requester, color)
			service.SendMessage(message.Channel(), reply)
			continue
		}

		if doesRoleHaveAuth(role.Permissions) {
			reply := fmt.Sprintf("Uh, %s, I think %s is more than just a colored role.", requester, color)
			service.SendMessage(message.Channel(), reply)
			continue
		}

		p.RolesByGuild[guildID].ManagedRoles[color] = true
	}

	printableRoles := p.getPrintableRoles(guildID)
	reply := fmt.Sprintf("Uh, I guess that means I am managing %v now.", printableRoles)
	service.SendMessage(message.Channel(), reply)
}

func (p *ColorPlugin) handleStopManaging(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage, guildID string) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	if message.UserID() != service.OwnerUserID {
		reply := fmt.Sprintf("Uh, %s, I think you need to ask my Rick for that command.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	_, parts := mmmorty.ParseCommand(service, message)

	if len(parts) == 1 {
		reply := fmt.Sprintf("Uh, %s, I think you forgot to name a color.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	for _, c := range parts[1:] {
		color := strings.ToLower(c)
		if !p.RolesByGuild[guildID].ManagedRoles[color] {
			reply := fmt.Sprintf("Uh, %s, I'm not managing %s", requester, color)
			service.SendMessage(message.Channel(), reply)
			continue
		}

		delete(p.RolesByGuild[guildID].ManagedRoles, color)
	}

	printableRoles := p.getPrintableRoles(guildID)
	reply := fmt.Sprintf("Uh, I guess that means I am managing %v now.", printableRoles)
	service.SendMessage(message.Channel(), reply)
}

// Save will save plugin state to a byte array.
func (p *ColorPlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

// Name returns the name of the plugin.
func (p *ColorPlugin) Name() string {
	return "Color"
}

// New will create a new Reminder plugin.
func New() mmmorty.Plugin {
	return &ColorPlugin{
		RolesByGuild: map[string]colorSet{},
	}
}
