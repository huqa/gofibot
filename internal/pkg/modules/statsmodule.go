package modules

import (
	"database/sql"
	"time"

	"github.com/huqa/gofibot/internal/pkg/logger"
	"github.com/lrstanley/girc"
)

const (
	statsTableStmt string = `
	CREATE TABLE IF NOT EXISTS stats_stats (
		channel text,
		nick text,
		hostmask text,
		words INT DEFAULT 0,
		PRIMARY KEY(channel, nick, hostmask)
	) WITHOUT ROWID;
	`
	upsertStmt string = `
	INSERT INTO stats_stats(channel, nick, hostmask, words) VALUES (?, ?, ?, ?) 
	ON CONFLICT(channel, nick, hostmask) DO UPDATE SET words = words + excluded.words;
	`
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
		scheduled: false,
		schedule:  time.Now(),
	}
}

// Init initializes Stats module
func (m *StatsModule) Init() error {
	m.log.Info("Init")

	_, err := m.db.Exec(statsTableStmt)
	if err != nil {
		m.log.Error("db error ", err)
		return err
	}

	return nil
}

// Stop is run when module is shut down
func (m *StatsModule) Stop() error {
	return nil
}

// Run Stats input to PRIVMSG target channel
func (m *StatsModule) Run(channel, hostmask, user, command, message string, args []string) error {
	//m.client.Cmd.Message(channel, user+": "+message)
	err := m.upsert(channel, user, hostmask, len(args))
	if err != nil {
		return err
	}

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

// Schedule returns true, time.Time if this module is scheduled to be run at time.Time
func (m *StatsModule) Schedule() (bool, time.Time) {
	return false, time.Time{}
}

func (m *StatsModule) upsert(channel, nick, hostmask string, words int) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(upsertStmt)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(channel, nick, hostmask, words)
	if err != nil {
		return err
	}
	return tx.Commit()
}