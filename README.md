# MMMorty!

Discord bot by me (aka Giant)

Much of this code was taken from:
https://github.com/iopred/bruxism/blob/master/discord.go

Check that one out. It's a really good example.

## Features

Mmmorty is rather Beta in the sense that it mainly has the features I felt like implementing
rather than any requested features.

#### TLDR

Use `@<botname> help` to view the commands.

#### Picking things

`@<botname> choose <option> or <option> (or ...)` - asks Morty to pick something for you.

If you want to opt out of this feature, start the bot with the `-pick=FALSE` command line flag.

#### Quotes

Use `@<botname> quote me` to view one of the stored quotes.

Quotes are stored by calling `@<botname> add quote <author> said <quote>`. For example:

<img src="docs/quotebot-example.png" height="350px">

There are hardcoded constants for the maximum word count of a quote as well as how many quotes can be stored, but these can be easily changed in the [Quote Plugin code](quoteplugin/quoteplugin.go). I may extract these into environment variables later.

Quotes with links (quotes containing the string "http") will be rejected.

If you want to opt out of this feature, start the bot with the `-quote=FALSE` command line flag.

#### Dictionary

Use `@<botname> define <word>` to get the definition of a word.

Words can be added by server staff with
`@<botname> add word <word> <definition>` and
deleted with `@<botname> delete <word>`.

#### Plot Twists

Plot twists work the same as Quotes except that they do not have an associated author. They are handy for users to submit ideas for arbitrary plot twists and prompts to help with storywriting. They are added with this:

    @<botname> add plot twist <some twist>

To retrieve twists:

    @<botname> plot twist

Twists are subject to the same restriction as quotes.

#### Setting Users' Colors With Roles

Use `@<botname> color me <color>` so mmmorty can set your color. This requires a bit of setup:

1. Create a set of roles that have no extra permissions applied (same as default for `@everyone`), with each role's name and color set as desired.
2. Call `@<botname> manage color <color list>`. For example, `@<botname> manage color red yellow green blue purple`
3. Make sure mmmorty's role is listed above the colors so it has permission to add/remove them.

To stop managing colors, use `@<botname> stop managing <color list>`. This could be handy either when removing/renaming a role or elevating its permissions and invalidating its use as a color-only role.

Mmmorty will refuse to assign roles which have any permissions applied or that are above it in the permissions list. It is expected, and recommended, to have colored roles function separately from user permissions.

If you want to opt out of this feature, start the bot with the `-color=FALSE` command line flag.

#### Timed Sprints

This is the original reason I made this bot.

Use `@<botname> start sprint at :XX for Y` to start a timed "sprint"/"word war".

This feature is targetted more towards the WriMo community but might be useful for anyone
looking for short bursts of productivity. Users can request to be pinged on these updates.
 
It should be easy to change the output strings in a fork branch to make use of the timer for other purposes.

How it works:

1. Start a sprint. The `:XX` notation provides a timezone-agnostic time to start.
  `Y` is the number of minutes the sprint should last. The `:` is optional and `at X` and `for Y` are interchangeable.
2. Users can use the `join sprint` and `leave sprint` commands to add/remove themselves from
  the list of users that get pinged at each interval.
  The user that starts the sprint is added automatically.
3. Users receive three updates: the one-minute-before warning, the start notification, and the end notification.

Multiple simultaneous sprints can be run. Each one is given an ID number to help manage them.
This feature is disabled by default, so you will need the `-war=TRUE` flag to enable it.

## Setting Up

1. Set up a bot with discord. A good guide for this is [here](https://github.com/reactiflux/discord-irc/wiki/Creating-a-discord-bot-&-getting-a-token).

2. Connect with permissions (see: https://discordapi.com/permissions.html). Use these permission:

    - Manage Roles (if you want the color-changing feature)
    - Read Messages
    - Send Messages

    This means you should connect to `https://discordapp.com/oauth2/authorize?client_id=<client id>&scope=bot&permissions=26843852868438528`

3. Run this bot. Use `make` to install mmmorty globally, and then run:

    `mmmorty -discordtoken <token> -discordowneruserid <your user id>`
    
  Alternatively you can set environment variables for `DISCORD_TOKEN` and `DISCORD_OWNER`
  so you only need to call `mmmorty` to run the program.

