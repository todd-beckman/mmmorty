package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/todd-beckman/mmmorty"
	"github.com/todd-beckman/mmmorty/colorplugin"
)

var discordToken string
var discordEmail string
var discordPassword string
var discordApplicationClientID string
var discordOwnerUserID string
var discordShards int

func init() {
	flag.StringVar(&discordToken, "discordtoken", "", "Discord token.")
	flag.StringVar(&discordEmail, "discordemail", "", "Discord account email.")
	flag.StringVar(&discordPassword, "discordpassword", "", "Discord account password.")
	flag.StringVar(&discordOwnerUserID, "discordowneruserid", "", "Discord owner user id.")
	flag.StringVar(&discordApplicationClientID, "discordapplicationclientid", "", "Discord application client id.")
	flag.IntVar(&discordShards, "discordshards", 1, "Number of discord shards.")
	flag.Parse()

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
			discord = mmmorty.NewDiscord(discordToken)
		} else {
			discord = mmmorty.NewDiscord(discordEmail, discordPassword)
		}
		discord.ApplicationClientID = discordApplicationClientID
		discord.OwnerUserID = discordOwnerUserID
		discord.Shards = discordShards
		bot.RegisterService(discord)

		bot.RegisterPlugin(discord, cp)
		bot.RegisterPlugin(discord, colorplugin.New())
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
