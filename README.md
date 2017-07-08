# MMMorty!

Discord bot in development.

Much of this code was taken from:
https://github.com/iopred/bruxism/blob/master/discord.go

Check that one out. It's a really good example.

## Usage

1. Set up a bot with discord. A good guide for this is [here](https://github.com/reactiflux/discord-irc/wiki/Creating-a-discord-bot-&-getting-a-token)

1. Connect with permissions (see: https://discordapi.com/permissions.html). When you install be sure to request permissions for at least:

    - Manage Roles
    - Read Messages
    - Send Messages

    This means you should connect to https://discordapp.com/oauth2/authorize?client_id=<client id>&scope=bot&permissions=26843852868438528

1. Run this bot. Use `make` to install mmmorty globally, and then run:

    `mmmorty -discordtoken <token> -discordapplicationclientid <client id>`
