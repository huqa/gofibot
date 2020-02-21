package modules

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/huqa/gofibot/internal/pkg/logger"
	"github.com/lrstanley/girc"
)

const (
	playerStatsTableStmt string = `
	CREATE TABLE IF NOT EXISTS guess_player_stats (
		nick text,
		guesses INT DEFAULT 0,
		rights INT DEFAULT 0,
		PRIMARY KEY(nick)
	) WITHOUT ROWID;
	`

	rollsTableStmt string = `
	CREATE TABLE IF NOT EXISTS guess_rolls (
		id INT,
		rolls INT DEFAULT 0,
		rights INT DEFAULT 0,
		PRIMARY KEY(id)
	);
	`

	upsertPlayerStatsStmt string = `
	INSERT INTO guess_player_stats(nick, guesses, rights) VALUES (?, ?, ?) 
	ON CONFLICT(nick) DO UPDATE SET guesses = guesses + excluded.guesses, rights = rights + excluded.rights;
	`
	upsertRollsStmt string = `
	INSERT INTO guess_rolls(id, rolls, rights) VALUES (?, ?, ?) 
	ON CONFLICT(id) DO UPDATE SET rolls = rolls + excluded.rolls, rights = rights + excluded.rights;
	`
	selectUserWordStat string = `
	SELECT guesses, rights FROM guess_player_stats WHERE nick = ?;
	`
)

// GuessModule is a guessing game
type GuessModule struct {
	*Module
	db *sql.DB
}

// NewGuessModule constructs a new GuessModule
func NewGuessModule(log logger.Logger, client *girc.Client, db *sql.DB) *GuessModule {
	return &GuessModule{
		&Module{
			log:      log.Named("guessmodule"),
			client:   client,
			global:   false,
			event:    "PRIVMSG",
			commands: []string{"arvaa"},
		},
		db,
	}
}

// Init initializes Guess module
func (m *GuessModule) Init() error {
	m.log.Info("Init")

	_, err := m.db.Exec(playerStatsTableStmt)
	if err != nil {
		m.log.Error("db error ", err)
		return err
	}

	_, err = m.db.Exec(rollsTableStmt)
	if err != nil {
		m.log.Error("db error ", err)
		return err
	}

	return nil
}

// Stop is run when module is shut down
func (m *GuessModule) Stop() error {
	return nil
}

// Run Stats input to PRIVMSG target channel
func (m *GuessModule) Run(channel, hostmask, user, command string, args []string) error {
	if len(args) == 0 {
		guesses, rights, err := m.getPlayerStats(user)
		if err != nil {
			m.client.Cmd.Message(channel, fmt.Sprintf("!arvaa - %s no bonus", user))
			return err
		}
		percent := (float64(rights) / float64(guesses)) * 100.0
		m.client.Cmd.Message(channel, fmt.Sprintf("!arvaa - %s olet arvannut %d kertaa joista %d on ollut oikein - onnistumisprosentti: %.2f", user, guesses, rights, percent))
		return nil
	}

	guess, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		m.client.Cmd.Message(channel, fmt.Sprintf("!arvaa - %s painu takas neukkulaan", user))
		return nil
	}

	if guess <= 0 || guess > 200 {
		m.client.Cmd.Message(channel, fmt.Sprintf("!arvaa - %s anna luku väliltä [1-200]", user))
		return nil
	}

	rand.Seed(time.Now().UnixNano())
	wasRight := false
	min := 1
	max := 200
	throw := rand.Intn(max-min+1) + min

	m.client.Cmd.Message(channel, fmt.Sprintf("!arvaa - %s arvasi %d ja heitti %d", user, guess, throw))

	if int64(throw) == guess {
		wasRight = true
		m.client.Cmd.Message(channel, fmt.Sprintf("!arvaa - %s CONGRATURATION YOU WINNER", user))
	}
	err = m.upsertRoll(throw, wasRight)
	if err != nil {
		m.log.Error("roll upsert error ", err)
	}
	err = m.upsertGuess(user, wasRight)
	if err != nil {
		m.log.Error("guess upsert error ", err)
	}
	return nil
}

// Commands returns commands used by this module
func (m *GuessModule) Commands() []string {
	return m.commands
}

// Event returns event type used by this module
func (m *GuessModule) Event() string {
	return m.event
}

// Global returns true if this module is a global command
func (m *GuessModule) Global() bool {
	return m.global
}

// Schedule returns true, time.Time if this module is scheduled to be run at time.Time
func (m *GuessModule) Schedule() (bool, time.Time, time.Duration) {
	return false, time.Time{}, 0
}

// upsert inserts or updates word counts on db
func (m *GuessModule) upsertGuess(nick string, wasRight bool) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(upsertPlayerStatsStmt)
	if err != nil {
		return err
	}
	defer stmt.Close()
	var wr = 0
	if wasRight == true {
		wr = 1
	}
	_, err = stmt.Exec(nick, 1, wr)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// upsert inserts or updates rolls on db
func (m *GuessModule) upsertRoll(number int, wasRight bool) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(upsertRollsStmt)
	if err != nil {
		return err
	}
	defer stmt.Close()
	var wr = 0
	if wasRight == true {
		wr = 1
	}
	_, err = stmt.Exec(number, 1, wr)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (m *GuessModule) getPlayerStats(nick string) (guesses int, rights int, err error) {
	tx, err := m.db.Begin()
	if err != nil {
		return 0, 0, err
	}
	stmt, err := tx.Prepare(selectUserWordStat)
	if err != nil {
		return 0, 0, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(nick).Scan(&guesses, &rights)
	if err != nil {
		return 0, 0, err
	}
	return guesses, rights, nil
}
