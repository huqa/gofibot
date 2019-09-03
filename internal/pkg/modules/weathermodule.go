package modules

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/huqa/gofibot/internal/pkg/logger"
	irc "github.com/thoj/go-ircevent"
)

type WeatherModule struct {
	log              logger.Logger
	public           bool
	event            string
	commands         []string
	conn             *irc.Connection
	weatherCollector *colly.Collector
	url              string
	weatherOptions   string
	wdResponses      map[string]*WeatherData
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
		event:          "PRIVMSG",
		public:         false,
		wdResponses:    make(map[string]*WeatherData, 0),
	}
}

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
		m.wdResponses[ID + Channel] = &WeatherData{
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
	//m.weatherCollector.Visit(weatherUrl)
	ID := strconv.FormatInt(time.Now().UnixNano(), 10)
	ctx := colly.NewContext()
	ctx.Put("ID", ID)
        ctx.Put("Channel", channel)
	m.weatherCollector.Request("GET", weatherUrl, nil, ctx, nil)
	m.weatherCollector.Wait()
	wd, ok := m.wdResponses[ID + channel]
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
	delete(m.wdResponses, ID + channel)
	m.conn.Privmsg(channel, wString)
	return nil
}

func (m *WeatherModule) Commands() []string {
	return m.commands
}

func (m *WeatherModule) Event() string {
	return m.event
}

func (m *WeatherModule) Public() bool {
	return m.public
}
