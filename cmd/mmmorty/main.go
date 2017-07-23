package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/todd-beckman/mmmorty"
	"github.com/todd-beckman/mmmorty/colorplugin"
	"github.com/todd-beckman/mmmorty/pickplugin"
	"github.com/todd-beckman/mmmorty/quoteplugin"
	"github.com/todd-beckman/mmmorty/warplugin"
)

var (
	discordToken               string
	discordEmail               string
	discordPassword            string
	discordApplicationClientID string
	discordOwnerUserID         string
	discordShards              int
	enableColor                bool
	enablePicking              bool
	enableQuotes               bool
	enableWars                 bool
)

const (
	OWNER_ENV = "DISCORD_OWNER"
	TOKEN_ENV = "DISCORD_TOKEN"
)

func init() {
	flag.StringVar(&discordToken, "discordtoken", "", "Discord token.")
	flag.StringVar(&discordEmail, "discordemail", "", "Discord account email.")
	flag.StringVar(&discordPassword, "discordpassword", "", "Discord account password.")
	flag.StringVar(&discordOwnerUserID, "discordowneruserid", "", "Discord owner user id.")
	flag.StringVar(&discordApplicationClientID, "discordapplicationclientid", "", "Discord application client id.")
	flag.IntVar(&discordShards, "discordshards", 1, "Number of discord shards.")
	flag.BoolVar(&enableColor, "color", true, "Whether to enable setting colors")
	flag.BoolVar(&enablePicking, "pick", true, "Whether to enable picking things")
	flag.BoolVar(&enableQuotes, "quote", true, "Whether to enable quoting people")
	flag.BoolVar(&enableWars, "war", false, "Whether to enable timed word wars")
	flag.Parse()

	if discordToken == "" {
		discordToken = os.Getenv(TOKEN_ENV)
	}
	if discordOwnerUserID == "" {
		discordOwnerUserID = os.Getenv(OWNER_ENV)
	}

	rand.Seed(time.Now().UnixNano())
}

func main() {
	q := make(chan bool)

	// Set our variables.
	bot := mmmorty.NewBot()

	// Generally CommandPlugins don't hold state, so we share one instance of the command plugin for all services.
	cp := mmmorty.NewCommandPlugin()

	cp.AddCommand("quit", func(bot *mmmorty.Bot, service mmmorty.Service, message mmmorty.Message, args string, parts []string) {
		if service.IsBotOwner(message) {
			q <- true
		}
	}, nil)

	// Register the Discord service if we have an email or token.
	if (discordEmail != "" && discordPassword != "") || discordToken != "" {
		var discord *mmmorty.Discord
		if discordToken != "" {
			discord = mmmorty.NewDiscord(fmt.Sprintf("Bot %s", discordToken))
		} else {
			discord = mmmorty.NewDiscord(discordEmail, discordPassword)
		}
		discord.ApplicationClientID = discordApplicationClientID
		discord.OwnerUserID = discordOwnerUserID
		discord.Shards = discordShards
		bot.RegisterService(discord)

		bot.RegisterPlugin(discord, cp)
		if enableColor {
			bot.RegisterPlugin(discord, colorplugin.New())
		}
		if enablePicking {
			bot.RegisterPlugin(discord, pickplugin.New())
		}
		if enableQuotes {
			bot.RegisterPlugin(discord, quoteplugin.New())
		}
		if enableWars {
			bot.RegisterPlugin(discord, warplugin.New())
		}
	} else {
		log.Println("(discordEmail and discordPassword) or discordToken is required.")
		os.Exit(1)
	}

	// Start all our services.
	bot.Open()

	// Wait for a termination signal, while saving the bot state every minute. Save on close.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	t := time.Tick(1 * time.Minute)

out:
	for {
		select {
		case <-q:
			break out
		case <-c:
			break out
		case <-t:
			bot.Save()
		}
	}

	bot.Save()
}
