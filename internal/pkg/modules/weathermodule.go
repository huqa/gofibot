package modules

import (
	"fmt"
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
	}
}

// Init initializes weather module
func (m *WeatherModule) Init() error {
	m.log.Info("Init")
	c := colly.NewCollector(
		colly.AllowedDomains("wttr.in"),
	)
	c.AllowURLRevisit = true
	c.OnResponse(m.weatherResponseCallback)
	c.OnError(func(r *colly.Response, err error) {
		m.log.Error("error: ", r.StatusCode, err)
                channel := r.Ctx.Get("Channel")
                m.client.Cmd.Message(channel, "!w - internet says: error no bonus")
	})
	m.weatherCollector = c
	return nil
}

// Stop is run when module is stopped
func (m *WeatherModule) Stop() error {
	return nil
}

// Run sends weather data to PRIVMSG target channel
func (m *WeatherModule) Run(channel, hostmask, user, command string, args []string) error {
	if len(args) == 0 {
		return nil
	}
        message := strings.Join(args, " ")
	weatherURL := fmt.Sprintf(m.url, message)
	weatherURL += m.weatherOptions
	ctx := colly.NewContext()
	ctx.Put("Channel", channel)
	m.weatherCollector.Request("GET", weatherURL, nil, ctx, nil)

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

// Schedule
func (m *WeatherModule) Schedule() (bool, time.Time, time.Duration) {
	return false, time.Time{}, 0
}

func (m *WeatherModule) weatherResponseCallback(r *colly.Response) {
	m.log.Debug("response received: ", r.StatusCode)
	channel := r.Ctx.Get("Channel")
	var body = string(r.Body)
	if strings.HasPrefix(body, "<html>") {
		m.client.Cmd.Message(channel, "!w - internet says: error no bonus")
		return
	}

	var wd = strings.Split(body, ",")
	data := &WeatherData{
		Location:      strings.Title(wd[0]),
		Description:   wd[1],
		Temperature:   wd[2],
		Humidity:      wd[3],
		Wind:          wd[4],
		Precipitation: wd[5],
	}

	wString := fmt.Sprintf(
		"!w - sää %s: %s - %s, ilmankosteus %s, tuuli %s, sademäärä %s",
		data.Location,
		data.Temperature,
		data.Description,
		data.Humidity,
		data.Wind,
		data.Precipitation,
	)
	m.client.Cmd.Message(channel, wString)
}
