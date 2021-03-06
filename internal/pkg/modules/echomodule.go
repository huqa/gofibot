package modules

import (
	"strings"
	"time"

	"github.com/huqa/gofibot/internal/pkg/logger"
	"github.com/lrstanley/girc"
)

// EchoModule echos input back to channel
type EchoModule struct {
	*Module
}

// NewEchoModule constructs new EchoModule
func NewEchoModule(log logger.Logger, client *girc.Client) *EchoModule {
	return &EchoModule{
		&Module{
			log:      log.Named("echomodule"),
			commands: []string{"echo"},
			client:   client,
			event:    "PRIVMSG",
		},
	}
}

// Init initializes echo module
func (m *EchoModule) Init() error {
	m.log.Info("Init")
	return nil
}

// Stop is run when module is stopped
func (m *EchoModule) Stop() error {
	return nil
}

// Run echos input to PRIVMSG target channel
func (m *EchoModule) Run(channel, hostmask, user, command string, args []string) error {
	message := strings.Join(args, " ")
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

// Schedule
func (m *EchoModule) Schedule() (bool, time.Time, time.Duration) {
	return false, time.Time{}, 0
}
