package modules

import (
	"database/sql"
	"fmt"
	"strings"
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
	top10WordStatsStmt string = `
	SELECT nick, hostmask, words FROM stats_stats WHERE channel = ? ORDER BY words DESC LIMIT 10;
	`
)

// StatsModule handles irc channel statistics
type StatsModule struct {
	*Module

	db *sql.DB

	scheduled bool
	schedule  time.Time
}

// NewStatsModule constructs a new StatsModule
func NewStatsModule(log logger.Logger, client *girc.Client, db *sql.DB) *StatsModule {
	return &StatsModule{
		&Module{
			log:      log.Named("Statsmodule"),
			client:   client,
			global:   true,
			event:    "PRIVMSG",
			commands: []string{"stats", "toptod"},
		},
		db,
		false,
		time.Time{},
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
	// handle global command -> upsert word count
	if command == "" {
		err := m.upsert(channel, user, hostmask, len(strings.Split(message, " ")))
		if err != nil {
			return err
		}
		return nil
	}
	output, err := m.selectWordStats(channel)
	if err != nil {
		m.log.Error("can't fetch word stats: ", err)
		return err
	}
	m.client.Cmd.Message(channel, output)
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
func (m *StatsModule) Schedule() (bool, time.Time, time.Duration) {
	return false, time.Time{}, 0
}

// upsert inserts or updates word counts on db
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

func (m *StatsModule) selectWordStats(channel string) (output string, err error) {
	stmt, err := m.db.Prepare(top10WordStatsStmt)
	if err != nil {
		return output, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(channel)
	if err != nil {
		return output, err
	}
	defer rows.Close()

	var (
		nick     string
		hostmask string
		words    int
		/*total    int
		mean     int
		median   int*/
	)

	i := 1
	output = ""
	for rows.Next() {
		err = rows.Scan(&nick, &hostmask, &words)
		if err != nil {
			return output, err
		}
		output += fmt.Sprintf("%d. %s(%d) ", i, nick, words)
		i++
	}
	return output, nil
}
