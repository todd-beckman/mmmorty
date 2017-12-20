package mmmorty

import (
	"errors"
)

// MessageType is a type used to determine the CRUD state of a message.
type MessageType string

const (
	// MessageTypeCreate is the message type for message creation.
	MessageTypeCreate MessageType = "create"
	// MessageTypeUpdate is the message type for message updates.
	MessageTypeUpdate = "update"
	// MessageTypeDelete is the message type for message deletion.
	MessageTypeDelete = "delete"
)

// ErrAlreadyJoined is an error dispatched on Join if the bot is already joined to the request.
var ErrAlreadyJoined = errors.New("already joined")

// LoadFunc is the function signature for a load handler.
type LoadFunc func(*Bot, Discord, []byte) error

// SaveFunc is the function signature for a save handler.
type SaveFunc func() ([]byte, error)

// HelpFunc is the function signature for a help handler.
type HelpFunc func(*Bot, Discord, DiscordMessage, bool) []string

// MessageFunc is the function signature for a message handler.
type MessageFunc func(*Bot, Discord, DiscordMessage)

// StatsFunc is the function signature for a stats handler.
type StatsFunc func(*Bot, Discord, DiscordMessage) []string

// Plugin is a plugin interface, supports loading and saving to a byte array and has help and message handlers.
type Plugin interface {
	Name() string
	Load(*Bot, Discord, []byte) error
	Save() ([]byte, error)
	Help(*Bot, Discord, DiscordMessage, bool) []string
	Message(*Bot, Discord, DiscordMessage)
}
