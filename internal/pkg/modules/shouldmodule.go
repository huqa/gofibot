package modules

import (
	"math/rand"
	"strings"
	"time"

	"github.com/huqa/gofibot/internal/pkg/logger"
	"github.com/lrstanley/girc"
)

// ShouldModule
type ShouldModule struct {
	*Module
	shoulds []string
}

// NewShouldModule
func NewShouldModule(log logger.Logger, client *girc.Client) *ShouldModule {
	return &ShouldModule{
		&Module{
			log:    log.Named("shouldmodule"),
			client: client,
			global: true,
			event:  "PRIVMSG",
		},
		[]string{
			"pitäiskö",
			"pitäskö",
			"pitäisikö",
			"pitääkö",
			"pitäsikö",
		},
	}
}

const (
	replyYes string = "pitäis"
	replyNo  string = "ei pitäis"
)

// Init initializes echo module
func (m *ShouldModule) Init() error {
	m.log.Info("Init")
	return nil
}

// Stop is run when module is stopped
func (m *ShouldModule) Stop() error {
	return nil
}

// Run
func (m *ShouldModule) Run(channel, hostmask, user, command string, args []string) error {
	message := strings.Join(args, " ")
	message = strings.ToLower(message)
	for _, sh := range m.shoulds {
		if strings.Contains(message, sh) {
			i := rand.Intn(11)
			if i >= 10 {
				m.client.Cmd.Message(channel, replyNo)
				return nil
			}
			m.client.Cmd.Message(channel, replyYes)
			return nil
		}
	}
	return nil
}

// Commands returns commands used by this module
func (m *ShouldModule) Commands() []string {
	return m.commands
}

// Event returns event type used by this module
func (m *ShouldModule) Event() string {
	return m.event
}

// Global returns true if this module is a global command
func (m *ShouldModule) Global() bool {
	return m.global
}

// Schedule
func (m *ShouldModule) Schedule() (bool, time.Time, time.Duration) {
	return false, time.Time{}, 0
}
