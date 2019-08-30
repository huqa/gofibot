package modules

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
	"github.com/huqa/gofibot/internal/pkg/logger"
	irc "github.com/thoj/go-ircevent"
)

type WeatherModule struct {
	log              logger.Logger
	commands         []string
	conn             *irc.Connection
	weatherCollector *colly.Collector
	url              string
	weatherOptions   string
	wd               WeatherData
}

type WeatherData struct {
	Location      string
	Description   string
	Temperature   string
	Humidity      string
	Wind          string
	Precipitation string
}

func NewWeatherModule(log logger.Logger, conn *irc.Connection) *WeatherModule {
	return &WeatherModule{
		log:            log.Named("weathermodule"),
		commands:       []string{"w", "sää", "saa"},
		conn:           conn,
		url:            "http://wttr.in/%s",
		weatherOptions: "?format=%l,%C,%t,%h,%w,%p&lang=fi",
	}
}

func (m *WeatherModule) Init() error {
	m.log.Info("Init")
	c := colly.NewCollector(
		colly.AllowedDomains("wttr.in"),
	)
	c.AllowURLRevisit = true
	// extract status code
	c.OnResponse(func(r *colly.Response) {
		m.log.Debug("response received: ", r.StatusCode)
		var wd = strings.Split(string(r.Body), ",")
		m.wd = WeatherData{
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

func (m *WeatherModule) Run(user, channel, message string, args []string) error {
	if len(args) == 0 {
		return nil
	}
	weatherUrl := fmt.Sprintf(m.url, args[0])
	weatherUrl += m.weatherOptions
	m.weatherCollector.Visit(weatherUrl)
	m.weatherCollector.Wait()
	wString := fmt.Sprintf("!w - sää %s: %s - %s, ilmankosteus %s, tuuli %s, sademäärä %s", m.wd.Location, m.wd.Temperature, m.wd.Description, m.wd.Humidity, m.wd.Wind, m.wd.Precipitation)
	m.conn.Privmsg(channel, wString)
	return nil
}

func (m *WeatherModule) Commands() []string {
	return m.commands
}
