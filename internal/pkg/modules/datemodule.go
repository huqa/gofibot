package modules

import (
	"fmt"
	"time"

	"github.com/huqa/gofibot/internal/pkg/logger"
	"github.com/lrstanley/girc"
)

// DateModule Dates input back to channel
type DateModule struct {
	*Module
	location *time.Location
}

const (
	dateString string = `Tänään on %s %s (viikko %d) vuoden %d. päivä.`
)

var ignoredChannels = map[string]int{}

// NewDateModule constructs new DateModule
func NewDateModule(log logger.Logger, client *girc.Client) *DateModule {
	return &DateModule{
		&Module{
			log:      log.Named("Datemodule"),
			commands: []string{"date", "pvm"},
			client:   client,
			event:    "PRIVMSG",
		},
		nil,
	}
}

// Init initializes Date module
func (m *DateModule) Init() error {
	m.log.Info("Init")
	loc, err := time.LoadLocation("Europe/Helsinki")
	if err != nil {
		m.log.Error("couldn't load time zone for helsinki: ", err)
		return err
	}
	m.location = loc
	return nil
}

// Stop is run when module is stopped
func (m *DateModule) Stop() error {
	return nil
}

// Run Dates input to PRIVMSG target channel
func (m *DateModule) Run(channel, hostmask, user, command, message string, args []string) error {
	if _, ok := ignoredChannels[channel]; ok {
		return nil
	}
	now := time.Now().In(m.location)
	weekday := m.finnishWeekday(now.Weekday().String())
	date := now.Format("2.1.2006")
	yearDay := now.YearDay()
	_, week := now.ISOWeek()

	output := fmt.Sprintf(dateString, weekday, date, week, yearDay)
	m.client.Cmd.Message(channel, output)
	return nil
}

// Commands returns commands used by this module
func (m *DateModule) Commands() []string {
	return m.commands
}

// Event returns event type used by this module
func (m *DateModule) Event() string {
	return m.event
}

// Global returns true if this module is a global command
func (m *DateModule) Global() bool {
	return m.global
}

// Schedule
func (m *DateModule) Schedule() (bool, time.Time, time.Duration) {
	/*dur, _ := time.ParseDuration("24h")
	t := time.Now()
	n := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, m.location)
	n = n.Add(dur)*/
	dur, _ := time.ParseDuration("2m")
	n := time.Now().Add(dur)
	return true, n, dur
}

func (m *DateModule) finnishWeekday(weekday string) string {
	switch weekday {
	case "Monday":
		return "Maanantai"
	case "Tuesday":
		return "Tiistai"
	case "Wednesday":
		return "Keskiviikko"
	case "Thursday":
		return "Torstai"
	case "Friday":
		return "Perjantai"
	case "Saturday":
		return "Lauantai"
	case "Sunday":
		return "Sunnuntai"
	default:
		return "Maanantai"
	}
}
