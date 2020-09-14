package modules

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/huqa/gofibot/internal/pkg/logger"
	"github.com/huqa/gofibot/internal/pkg/utils"
	"github.com/lrstanley/girc"
	bolt "go.etcd.io/bbolt"
)

const (
	guessRootBucket  string = `Guess`
	guessStatsBucket string = `Stats`
	guessRollsBucket string = `Rolls`

	guessLimitPerDay int = 5

	statsCommand string = "arvaa-stats"
)

// Guess represents a guess from a player
type Guess struct {
	Nick    string
	Guesses int
	Rights  int
}

// Roll represents guessed dice values
type Roll struct {
	Value  int
	Rolls  int
	Rights int
}

// GuessModule is a guessing game
type GuessModule struct {
	*Module
	location     *time.Location
	db           *bolt.DB
	guessesToday map[string]int
}

// NewGuessModule constructs a new GuessModule
func NewGuessModule(log logger.Logger, client *girc.Client, db *bolt.DB) *GuessModule {
	return &GuessModule{
		&Module{
			log:      log.Named("guessmodule"),
			client:   client,
			global:   false,
			event:    "PRIVMSG",
			commands: []string{"arvaa", statsCommand},
		},
		nil,
		db,
		make(map[string]int),
	}
}

// Init initializes Guess module
func (m *GuessModule) Init() error {
	m.log.Info("Init")

	loc, err := time.LoadLocation("Europe/Helsinki")
	if err != nil {
		m.log.Error("couldn't load time zone for helsinki: ", err)
		return err
	}
	m.location = loc

	err = m.db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte(guessRootBucket))
		if err != nil {
			return fmt.Errorf("could not create root bucket: %v", err)
		}
		_, err = root.CreateBucketIfNotExists([]byte(guessStatsBucket))
		if err != nil {
			return fmt.Errorf("could not create weight bucket: %v", err)
		}
		_, err = root.CreateBucketIfNotExists([]byte(guessRollsBucket))
		if err != nil {
			return fmt.Errorf("could not create days bucket: %v", err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("could not set up buckets, %v", err)
	}

	return nil
}

// Stop is run when module is shut down
func (m *GuessModule) Stop() error {
	return nil
}

// Run Stats input to PRIVMSG target channel
func (m *GuessModule) Run(channel, hostmask, user, command string, args []string) error {
	if command == statsCommand {
		rolls, _, err := m.getRollStats()
		if err != nil {
			m.Module.log.Error("roll stats error", err)
		}
		/*allRolls := make(map[int]int, 0)
		allRights := make(map[int]int, 0)
		for k, roll := range rolls {
			allRolls[k] = roll.Rolls
			allRights[k] = roll.Rights
		}
		for k, r := range allRolls {

		}*/
		allRights := make([]Roll, len(rolls))
		copy(allRights, rolls)
		// sort in reverse order
		sort.Slice(rolls, func(i, j int) bool {
			return rolls[i].Rolls > rolls[j].Rolls
		})
		sort.Slice(allRights, func(i, j int) bool {
			return allRights[i].Rights > allRights[j].Rights
		})

		i := 1
		output := ""
		for _, r := range rolls {
			if i < 11 {
				output += fmt.Sprintf("(%d:%d) ", r.Value, r.Rolls)
			}
			i++
		}

		i = 1
		output1 := ""
		for _, r := range allRights {
			if i < 11 {
				output1 += fmt.Sprintf("(%d:%d) ", r.Value, r.Rights)
			}
			i++
		}
		m.client.Cmd.Message(channel, "!arvaa-stats TOP 10 heitetyt luvut (nopanluku:määrä)")
		m.client.Cmd.Message(channel, output)

		m.client.Cmd.Message(channel, "!arvaa-stats TOP 10 oikein arvatut luvut (nopanluku:määrä)")
		m.client.Cmd.Message(channel, output1)

		return nil
	}
	if hostmask == "SYSTEM" && user == "SYSTEM" {
		m.resetDailyGuesses()
		return nil
	}
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
		m.client.Cmd.Message(channel, fmt.Sprintf("!arvaa - %s anna luku väliltä 1-200", user))
		return nil
	}

	guessesLeft := m.handleGuessLimit(user)
	if guessesLeft < 0 {
		m.client.Cmd.Message(channel, fmt.Sprintf("!arvaa - %s ei arvauksia jäljellä tänään", user))
		return nil
	}

	rand.Seed(time.Now().UnixNano())
	wasRight := false
	min := 1
	max := 200
	throw := rand.Intn(max-min+1) + min

	m.client.Cmd.Message(channel, fmt.Sprintf("!arvaa - %s arvasi %d ja heitti %d - arvauksia jäljellä %d", user, guess, throw, guessesLeft))

	if int64(throw) == guess {
		wasRight = true
		m.client.Cmd.Message(channel, fmt.Sprintf("!arvaa - %s CONGRATURATIONS YOU WINRAR", user))
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
	dur, _ := time.ParseDuration("24h")
	t := time.Now()
	n := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, m.location)
	n = n.Add(dur)
	return true, n, dur
}

func (m *GuessModule) handleGuessLimit(nick string) (guessesLeft int) {
	guessesLeft = 0
	if guesses, ok := m.guessesToday[nick]; ok {
		guessesLeft = guesses - 1
		if guesses < 0 {
			return -1
		}
		m.guessesToday[nick] = guessesLeft
		return guessesLeft
	}

	guessesLeft = guessLimitPerDay - 1
	m.guessesToday[nick] = guessLimitPerDay - 1
	return guessesLeft
}

func (m *GuessModule) resetDailyGuesses() {
	m.guessesToday = make(map[string]int)
}

// upsert inserts or updates guesses count to db
func (m *GuessModule) upsertGuess(nick string, wasRight bool) error {
	var wr = 0
	if wasRight == true {
		wr = 1
	}
	return m.db.Update(func(tx *bolt.Tx) error {
		guessBucket := tx.Bucket([]byte(guessRootBucket)).Bucket([]byte(guessStatsBucket))
		guessBytes := guessBucket.Get([]byte(nick))
		if guessBytes == nil {
			insert := Guess{
				Nick:    nick,
				Guesses: 1,
				Rights:  wr,
			}
			enc, err := json.Marshal(insert)
			if err != nil {
				return err
			}
			return guessBucket.Put([]byte(nick), enc)
		}
		var g Guess
		err := json.Unmarshal(guessBytes, &g)
		if err != nil {
			return err
		}
		g.Guesses = g.Guesses + 1
		g.Rights = g.Rights + wr
		enc, err := json.Marshal(g)
		if err != nil {
			return err
		}
		return guessBucket.Put([]byte(nick), enc)
	})
}

// upsert inserts or updates rolls on db
func (m *GuessModule) upsertRoll(number int, wasRight bool) error {
	var wr = 0
	if wasRight == true {
		wr = 1
	}
	return m.db.Update(func(tx *bolt.Tx) error {
		rollsBucket := tx.Bucket([]byte(guessRootBucket)).Bucket([]byte(guessRollsBucket))
		rollBytes := rollsBucket.Get(utils.Itob(number))
		if rollBytes == nil {
			insert := Roll{
				Value:  number,
				Rolls:  1,
				Rights: wr,
			}
			enc, err := json.Marshal(insert)
			if err != nil {
				return err
			}
			return rollsBucket.Put(utils.Itob(number), enc)
		}
		var r Roll
		err := json.Unmarshal(rollBytes, &r)
		if err != nil {
			return err
		}
		r.Rolls = r.Rolls + 1
		r.Rights = r.Rights + wr
		enc, err := json.Marshal(r)
		if err != nil {
			return err
		}
		return rollsBucket.Put(utils.Itob(number), enc)
	})
}

func (m *GuessModule) getPlayerStats(nick string) (guesses int, rights int, err error) {
	err = m.db.View(func(tx *bolt.Tx) error {
		guessBucket := tx.Bucket([]byte(guessRootBucket)).Bucket([]byte(guessStatsBucket))
		guessBytes := guessBucket.Get([]byte(nick))
		if guessBytes == nil {
			guesses = 0
			rights = 0
			return errors.New("no user data found")
		}
		var g Guess
		err := json.Unmarshal(guessBytes, &g)
		if err != nil {
			return err
		}
		guesses = g.Guesses
		rights = g.Rights
		return nil
	})
	return guesses, rights, nil
}

func (m *GuessModule) getRollStats() (rolls []Roll, keys []int, err error) {
	rolls = make([]Roll, 0)
	keys = make([]int, 0)
	err = m.db.View(func(tx *bolt.Tx) error {
		rollBucket := tx.Bucket([]byte(guessRootBucket)).Bucket([]byte(guessRollsBucket))
		return rollBucket.ForEach(func(k, v []byte) error {
			var key int
			key = utils.Btoi(k)
			keys = append(keys, key)
			var r Roll
			err := json.Unmarshal(v, &r)
			if err != nil {
				return err
			}
			rolls = append(rolls, r)
			return nil
		})
	})
	if err != nil {
		return rolls, keys, err
	}
	return rolls, keys, nil
}
