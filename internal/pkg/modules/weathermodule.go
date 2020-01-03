package modules

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/huqa/gofibot/internal/pkg/logger"
	"github.com/lrstanley/girc"
)

// WeatherModule fetches weather from an outside service
type WeatherModule struct {
	*Module
	weatherCollector *colly.Collector
	url              string
	weatherOptions   string
	wdResponses      map[string]*WeatherData
}

// WeatherData defines the basic data structure for weather data
type WeatherData struct {
	Location      string
	Description   string
	Temperature   string
	Humidity      string
	Wind          string
	Precipitation string
}

// NewWeatherModule constructs new WeatherModule
func NewWeatherModule(log logger.Logger, client *girc.Client) *WeatherModule {
	return &WeatherModule{
		&Module{
			log:      log.Named("weathermodule"),
			commands: []string{"w", "sää", "saa"},
			client:   client,
			event:    "PRIVMSG",
		},
		nil,
		"http://wttr.in/%s",
		"?format=%l,%C,%t,%h,%w,%p&lang=fi",
		make(map[string]*WeatherData, 0),
	}
}

// Init initializes weather module
func (m *WeatherModule) Init() error {
	m.log.Info("Init")
	c := colly.NewCollector(
		colly.AllowedDomains("wttr.in"),
	)
	c.AllowURLRevisit = true
	c.OnResponse(func(r *colly.Response) {
		m.log.Debug("response received: ", r.StatusCode)
		var wd = strings.Split(string(r.Body), ",")
		ID := r.Ctx.Get("ID")
		Channel := r.Ctx.Get("Channel")
		m.wdResponses[ID+Channel] = &WeatherData{
			Location:      strings.Title(wd[0]),
			Description:   wd[1],
			Temperature:   wd[2],
			Humidity:      wd[3],
			Wind:          wd[4],
			Precipitation: wd[5],
		}
	})
	c.OnError(func(r *colly.Response, err error) {
		m.log.Error("error: ", r.StatusCode, err)
	})
	m.weatherCollector = c
	return nil
}

// Stop is run when module is stopped
func (m *WeatherModule) Stop() error {
	return nil
}

// Run sends weather data to PRIVMSG target channel
func (m *WeatherModule) Run(channel, hostmask, user, command, message string, args []string) error {
	if len(args) == 0 {
		return nil
	}
	weatherURL := fmt.Sprintf(m.url, message)
	weatherURL += m.weatherOptions
	ID := strconv.FormatInt(time.Now().UnixNano(), 10)
	ctx := colly.NewContext()
	ctx.Put("ID", ID)
	ctx.Put("Channel", channel)
	key := ID + channel
	m.weatherCollector.Request("GET", weatherURL, nil, ctx, nil)
	m.weatherCollector.Wait()
	wd, ok := m.wdResponses[key]
	if !ok {
		return nil
	}
	wString := fmt.Sprintf(
		"!w - sää %s: %s - %s, ilmankosteus %s, tuuli %s, sademäärä %s",
		wd.Location,
		wd.Temperature,
		wd.Description,
		wd.Humidity,
		wd.Wind,
		wd.Precipitation,
	)
	delete(m.wdResponses, key)
	m.client.Cmd.Message(channel, wString)
	return nil
}

// Commands return all commands used by the module
func (m *WeatherModule) Commands() []string {
	return m.commands
}

// Event returns event type used by this module
func (m *WeatherModule) Event() string {
	return m.event
}

// Global returns true if this module is a global command
func (m *WeatherModule) Global() bool {
	return m.global
}

func (m *WeatherModule) Schedule() (bool, time.Time) {
	return false, time.Time{}
}
