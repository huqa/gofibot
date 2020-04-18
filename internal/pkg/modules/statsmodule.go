package modules

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/huqa/gofibot/internal/pkg/logger"
	"github.com/lrstanley/girc"
	bolt "go.etcd.io/bbolt"
)

const (
	statsRootBucket    string = "Stats"
	channelStatsBucket string = "Channel"
)

// ChannelStats represents a users chat stats on a channel
type ChannelStats struct {
	Nick     string
	Channel  string
	Hostmask string
	Words    int
}

// StatsModule handles irc channel statistics
type StatsModule struct {
	*Module

	db *bolt.DB

	location *time.Location
}

// NewStatsModule constructs a new StatsModule
func NewStatsModule(log logger.Logger, client *girc.Client, db *bolt.DB) *StatsModule {
	return &StatsModule{
		&Module{
			log:      log.Named("Statsmodule"),
			client:   client,
			global:   true,
			event:    "PRIVMSG",
			commands: []string{"stats", "toptod"},
		},
		db,
		nil,
	}
}

// Init initializes Stats module
func (m *StatsModule) Init() error {
	m.log.Info("Init")

	err := m.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(statsRootBucket))
		if err != nil {
			return fmt.Errorf("could not create root bucket: %v", err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("could not set up buckets, %v", err)
	}

	loc, err := time.LoadLocation("Europe/Helsinki")
	if err != nil {
		m.log.Error("couldn't load time zone for helsinki: ", err)
		return err
	}
	m.location = loc

	return nil
}

// Stop is run when module is shut down
func (m *StatsModule) Stop() error {
	return nil
}

// Run Stats input to PRIVMSG target channel
func (m *StatsModule) Run(channel, hostmask, user, command string, args []string) error {
	// handle global command -> upsert word count
	if command == "" && hostmask != "SYSTEM" {
		err := m.upsert(channel, user, hostmask, len(args))
		if err != nil {
			m.log.Error("upsert error: ", err)
			return err
		}
		return nil
	}
	output, output2, err := m.selectWordStats(channel)
	if err != nil {
		m.log.Error("can't fetch word stats: ", err)
		return err
	}
	m.client.Cmd.Message(channel, output)
	m.client.Cmd.Message(channel, output2)
	if hostmask == "SYSTEM" && user == "SYSTEM" {
		err = m.clearStats(channel)
		if err != nil {
			m.log.Error("can't clear word stats: ", err)
		}
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
func (m *StatsModule) Schedule() (bool, time.Time, time.Duration) {
	dur, _ := time.ParseDuration("24h")
	t := time.Now()
	n := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, m.location)
	n = n.Add(dur)
	return true, n, dur
}

func (m *StatsModule) clearStats(channel string) error {
	return m.db.Update(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte(statsRootBucket))
		chanBucket := root.Bucket([]byte(channel))
		if chanBucket == nil {
			return nil
		}
		return root.DeleteBucket([]byte(channel))
	})
}

// upsert inserts or updates word counts on db
func (m *StatsModule) upsert(channel, nick, hostmask string, words int) error {
	return m.db.Update(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte(statsRootBucket))
		chanBucket, err := root.CreateBucketIfNotExists([]byte(channel))
		if err != nil {
			return err
		}
		statsBytes := chanBucket.Get([]byte(nick))
		if statsBytes == nil {
			insert := ChannelStats{
				Nick:     nick,
				Channel:  channel,
				Hostmask: hostmask,
				Words:    words,
			}
			enc, err := json.Marshal(insert)
			if err != nil {
				return err
			}
			return chanBucket.Put([]byte(nick), enc)
		}
		var c ChannelStats
		err = json.Unmarshal(statsBytes, &c)
		if err != nil {
			return err
		}
		c.Words = c.Words + words
		enc, err := json.Marshal(c)
		if err != nil {
			return err
		}
		return chanBucket.Put([]byte(nick), enc)
	})
}

func (m *StatsModule) selectWordStats(channel string) (output string, output2 string, err error) {
	stats := make([]ChannelStats, 0)
	err = m.db.View(func(tx *bolt.Tx) error {
		chanBucket := tx.Bucket([]byte(statsRootBucket)).Bucket([]byte(channel))
		if chanBucket == nil {
			return nil
		}
		c := chanBucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var cs ChannelStats
			err = json.Unmarshal(v, &cs)
			if err != nil {
				continue
			}
			stats = append(stats, cs)
		}
		return nil
	})
	if err != nil {
		return "", "", err
	}
	if len(stats) <= 0 {
		return "", "", nil
	}
	// sort in reverse order
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Words > stats[j].Words
	})

	var (
		total int
		mean  int
	)

	i := 1
	output = ""
	for _, cs := range stats {
		total += cs.Words
		if i < 11 {
			output += fmt.Sprintf("%d. %s(%d) ", i, cs.Nick, cs.Words)
		}
		i++
	}
	mean = total / (i - 1)
	output2 = fmt.Sprintf("Kaikki yhteensä: %d Keskiarvo: %d", total, mean)
	return output, output2, nil
}
