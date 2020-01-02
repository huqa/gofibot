package modules

import (
	"database/sql"
	"time"

	"github.com/huqa/gofibot/internal/pkg/logger"
	"github.com/lrstanley/girc"
)

// StatsModule handles irc channel statistics
type StatsModule struct {
	log      logger.Logger
	commands []string
	client   *girc.Client
	event    string
	global   bool

	scheduled bool
	schedule  time.Time

	db *sql.DB
}

// NewStatsModule constructs a new StatsModule
func NewStatsModule(log logger.Logger, client *girc.Client, db *sql.DB) *StatsModule {
	return &StatsModule{
		log:       log.Named("Statsmodule"),
		commands:  []string{"toptod", "stats"},
		client:    client,
		event:     "PRIVMSG",
		global:    false,
		db:        db,
		scheduled: true,
		schedule:  time.Now(),
	}
}

// Init initializes Stats module
func (m *StatsModule) Init() error {
	m.log.Info("Init")
	return nil
}

// Run Stats input to PRIVMSG target channel
func (m *StatsModule) Run(user, channel, message string, args []string) error {
	m.client.Cmd.Message(channel, user+": "+message)
	return nil
}

// Commands returns commands used by this module
func (m *StatsModule) Commands() []string {
	return m.commands
}

// Event returns event type used by this module
func (m *StatsModule) Event() string {
	return m.event
}

// Global returns true if this module is a global command
func (m *StatsModule) Global() bool {
	return m.global
}

func (m *StatsModule) Schedule() (bool, time.Time) {
	return false, time.Time{}
}
