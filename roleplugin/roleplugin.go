package roleplugin

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/todd-beckman/mmmorty"
)

const rolesCommand = "i am"
const manageRolesCommand = "managerole"
const stopManagingCommand = "stopmanagingrole"

type rolesSet struct {
	ManagedRoles map[string]bool `json:"managedRoles"`
}

// RolePlugin is the save data for this plugin
type RolePlugin struct {
	bot          *mmmorty.Bot
	RolesByGuild map[string]rolesSet `json:"rolesByGuild"`
}

// Used to determine if the role is more than aesthetic
var authPermissions int64 = 0x00000002 | // kick
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

func doesRoleHaveAuth(permissions int64) bool {
	return permissions&authPermissions > 0
}

func (p *RolePlugin) getPrintableRoles(guildID string) []string {
	printableRoles := []string{}
	for role, isManaged := range p.RolesByGuild[guildID].ManagedRoles {
		if isManaged {
			printableRoles = append(printableRoles, role)
		}
	}
	return printableRoles
}

// Help gets the usage for this plugin
func (p *RolePlugin) Help(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage, detailed bool) []string {
	help := mmmorty.CommandHelp(service, rolesCommand, "role", "assigns the desired role if this server supports it.")
	return help
}

// Load loads this plugin from the given data
func (p *RolePlugin) Load(bot *mmmorty.Bot, service mmmorty.Discord, data []byte) error {
	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
			return err
		}
	}

	return nil
}

// Message is the command handler for this plugin
func (p *RolePlugin) Message(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage) {
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
		p.RolesByGuild = map[string]rolesSet{
			guildID: rolesSet{},
		}
	}

	if p.RolesByGuild[guildID].ManagedRoles == nil {
		p.RolesByGuild[guildID] = rolesSet{
			ManagedRoles: map[string]bool{},
		}
	}

	if mmmorty.MatchesCommand(service, rolesCommand, message) {
		p.handleIAm(bot, service, message, guildID)
	} else if mmmorty.MatchesCommand(service, manageRolesCommand, message) {
		p.handleManageRole(bot, service, message, guildID)
	} else if mmmorty.MatchesCommand(service, stopManagingCommand, message) {
		p.handleStopManaging(bot, service, message, guildID)
	}
}

func (p *RolePlugin) handleIAm(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage, guildID string) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	if service.IsPrivate(message) {
		reply := fmt.Sprintf("Uh, %s, I cannot assign roles in private.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	if availableRoles := p.getPrintableRoles(guildID); len(availableRoles) == 0 {
		reply := fmt.Sprintf("Uh, %s, I don't think this server lets me set that role.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	_, parts := mmmorty.ParseCommand(service, message)

	if len(parts) < 1 {
		reply := fmt.Sprintf("Uh, %s, I think you forgot to name a role.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	for _, roleName := range parts {
		role := service.GetRoleByName(message.Channel(), roleName)
		if role == nil {
			reply := fmt.Sprintf("Uh, %s, I can't find a role called %s", requester, roleName)
			service.SendMessage(message.Channel(), reply)
			return
		}

		if doesRoleHaveAuth(role.Permissions) {
			reply := fmt.Sprintf("Uh, %s, I'm not supposed to share that role.", requester)
			service.SendMessage(message.Channel(), reply)
			return
		}

		ok := service.GuildMemberRoleAdd(guildID, message.UserID(), role.ID)
		if !ok {
			reply := fmt.Sprintf("Uh, %s, something went wrong. Are you sure I can let you be %v?", requester, roleName)
			service.SendMessage(message.Channel(), reply)
			return
		}

		reply := fmt.Sprintf("You got it, %s! You are now %s", requester, roleName)
		service.SendMessage(message.Channel(), reply)
	}
	return

}

func (p *RolePlugin) handleManageRole(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage, guildID string) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	if service.IsPrivate(message) {
		reply := fmt.Sprintf("Uh, %s, I cannot manage roles in private.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	if message.UserID() != service.OwnerUserID {
		reply := fmt.Sprintf("Uh, %s, I think you need to ask my Rick for that command.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	_, parts := mmmorty.ParseCommand(service, message)

	if len(parts) < 1 {
		reply := fmt.Sprintf("Uh, %s, I think you forgot to name a role.", requester)

		service.SendMessage(message.Channel(), reply)
		return
	}

	for _, c := range parts {
		roleName := strings.ToLower(c)
		if p.RolesByGuild[guildID].ManagedRoles[roleName] {
			reply := fmt.Sprintf("Uh, %s, I am already managing %s", requester, roleName)
			service.SendMessage(message.Channel(), reply)
			continue
		}

		role := service.GetRoleByName(message.Channel(), roleName)
		if role == nil {
			reply := fmt.Sprintf("Uh, %s, I can't find a role called %s", requester, roleName)
			service.SendMessage(message.Channel(), reply)
			continue
		}

		if doesRoleHaveAuth(role.Permissions) {
			reply := fmt.Sprintf("Uh, %s, I don't think I can manage that role.", requester, roleName)
			service.SendMessage(message.Channel(), reply)
			continue
		}

		p.RolesByGuild[guildID].ManagedRoles[roleName] = true
	}

	printableRoles := p.getPrintableRoles(guildID)
	reply := fmt.Sprintf("Uh, I guess that means I am managing %v now.", printableRoles)
	service.SendMessage(message.Channel(), reply)
}

func (p *RolePlugin) handleStopManaging(bot *mmmorty.Bot, service mmmorty.Discord, message mmmorty.DiscordMessage, guildID string) {
	requester := fmt.Sprintf("<@%s>", message.UserID())

	if message.UserID() != service.OwnerUserID {
		reply := fmt.Sprintf("Uh, %s, I think you need to ask my Rick for that command.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	_, parts := mmmorty.ParseCommand(service, message)

	if len(parts) < 1 {
		reply := fmt.Sprintf("Uh, %s, I think you forgot to name a role.", requester)
		service.SendMessage(message.Channel(), reply)
		return
	}

	for _, c := range parts {
		role := strings.ToLower(c)
		if !p.RolesByGuild[guildID].ManagedRoles[role] {
			reply := fmt.Sprintf("Uh, %s, I'm not managing %s", requester, role)
			service.SendMessage(message.Channel(), reply)
			continue
		}

		delete(p.RolesByGuild[guildID].ManagedRoles, role)
	}

	printableRoles := p.getPrintableRoles(guildID)
	reply := fmt.Sprintf("Uh, I guess that means I am managing %v now.", printableRoles)
	service.SendMessage(message.Channel(), reply)
}

// Save will save plugin state to a byte array.
func (p *RolePlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

// Name returns the name of the plugin.
func (p *RolePlugin) Name() string {
	return "Roles"
}

// New will create a new Reminder plugin.
func New() mmmorty.Plugin {
	return &RolePlugin{
		RolesByGuild: map[string]rolesSet{},
	}
}
