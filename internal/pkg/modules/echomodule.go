package modules

import (
	"github.com/huqa/gofibot/internal/pkg/logger"
	"github.com/lrstanley/girc"
)

// EchoModule echos input back to channel
type EchoModule struct {
	log      logger.Logger
	commands []string
	client   *girc.Client
	event    string
	global   bool
}

// NewEchoModule constructs new EchoModule
func NewEchoModule(log logger.Logger, client *girc.Client) *EchoModule {
	return &EchoModule{
		log:      log.Named("echomodule"),
		commands: []string{"echo"},
		client:   client,
		event:    "PRIVMSG",
		global:   false,
	}
}

// Init initializes echo module
func (m *EchoModule) Init() error {
	m.log.Info("Init")
	return nil
}

// Run echos input to PRIVMSG target channel
func (m *EchoModule) Run(user, channel, message string, args []string) error {
	m.client.Cmd.Message(channel, user+": "+message)
	return nil
}

// Commands returns commands used by this module
func (m *EchoModule) Commands() []string {
	return m.commands
}

// Event returns event type used by this module
func (m *EchoModule) Event() string {
	return m.event
}

// Global returns true if this module is a global command
func (m *EchoModule) Global() bool {
	return m.global
}
