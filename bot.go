package mmmorty

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime/debug"
)

// VersionString is the current version of the bot
const VersionString string = "0.11"

type serviceEntry struct {
	Discord
	Plugins         map[string]Plugin
	messageChannels []chan DiscordMessage
}

// Bot enables registering of Services and Plugins.
type Bot struct {
	Services    map[string]*serviceEntry
	ImgurID     string
	ImgurAlbum  string
	MashableKey string
}

// MessageRecover is the default panic handler
func (b *Bot) MessageRecover(discord Discord, channel string) {
	if r := recover(); r != nil {
		panic := fmt.Sprintf("%s", r)
		// log first
		log.Println(panic)
		log.Println("Recovered:", string(debug.Stack()))

		// notify owner
		owner := fmt.Sprintf("<@%s>", discord.OwnerUserID)
		discord.SendMessage(channel, fmt.Sprintf("%s: Something went wrong. Summary: %s", owner, panic))
	}
}

// NewBot will create a new bot.
func NewBot() *Bot {
	return &Bot{
		Services: make(map[string]*serviceEntry, 0),
	}
}

func (b *Bot) getData(service Discord, plugin Plugin) []byte {
	if b, err := ioutil.ReadFile(service.Name() + "/" + plugin.Name()); err == nil {
		return b
	}
	return nil
}

// RegisterService registers a service with the bot.
func (b *Bot) RegisterService(service Discord) {
	if b.Services[service.Name()] != nil {
		log.Println("Service with that name already registered", service.Name())
	}
	serviceName := service.Name()
	b.Services[serviceName] = &serviceEntry{
		Discord: service,
		Plugins: make(map[string]Plugin, 0),
	}
	b.RegisterPlugin(service, NewHelpPlugin())
}

// RegisterPlugin registers a plugin on a service.
func (b *Bot) RegisterPlugin(service Discord, plugin Plugin) {
	s := b.Services[service.Name()]
	if s.Plugins[plugin.Name()] != nil {
		log.Println("Plugin with that name already registered", plugin.Name())
	}
	s.Plugins[plugin.Name()] = plugin
}

func (b *Bot) listen(service Discord, messageChan <-chan *DiscordMessage) {
	serviceName := service.Name()

	for {
		message := <-messageChan
		//log.Printf("<%s> %s: %s\n", message.Channel(), message.UserName(), message.Message())
		plugins := b.Services[serviceName].Plugins
		for _, plugin := range plugins {
			go plugin.Message(b, service, *message)
		}
	}
}

// Open will open all the current services and begins listening.
func (b *Bot) Open() {
	for _, service := range b.Services {
		if messageChan, err := service.Open(); err == nil {
			for _, plugin := range service.Plugins {
				plugin.Load(b, service.Discord, b.getData(service.Discord, plugin))
			}
			go b.listen(service.Discord, messageChan)
		} else {
			log.Printf("Error creating service %s: %v\n", service.Name(), err)
		}
	}
}

// Save will save the current plugin state for all plugins on all services.
func (b *Bot) Save() {
	for _, service := range b.Services {
		serviceName := service.Name()
		if err := os.Mkdir(serviceName, os.ModePerm); err != nil {
			if !os.IsExist(err) {
				log.Println("Error creating service directory.")
			}
		}
		for _, plugin := range service.Plugins {
			if data, err := plugin.Save(); err != nil {
				log.Printf("Error saving plugin %s %s. %v", serviceName, plugin.Name(), err)
			} else if data != nil {
				if err := ioutil.WriteFile(serviceName+"/"+plugin.Name(), data, os.ModePerm); err != nil {
					log.Printf("Error saving plugin %s %s. %v", serviceName, plugin.Name(), err)
				}
			}
		}
	}
}
