package modules

import (
	"net/url"

	"github.com/gocolly/colly"
	"github.com/huqa/gofibot/internal/pkg/logger"
	irc "github.com/thoj/go-ircevent"
)

type URLTitleModule struct {
	log            logger.Logger
	commands       []string
	conn           *irc.Connection
	titleCollector *colly.Collector
	event          string
	public         bool
}

type URLTitle string

func NewURLTitleModule(log logger.Logger, conn *irc.Connection) *URLTitleModule {
	return &URLTitleModule{
		log:      log.Named("urltitlemodule"),
		commands: []string{},
		conn:     conn,
		event:    "PRIVMSG",
		public:   true,
	}
}

func (m *URLTitleModule) Init() error {
	m.log.Info("Init")
	c := colly.NewCollector(
		colly.Async(true),
		colly.AllowURLRevisit(),
	)
	c.AllowURLRevisit = true
	c.OnHTML("title", func(e *colly.HTMLElement) {

	})
	return nil
}

func (m *URLTitleModule) Run(user, channel, message string, args []string) error {
	//m.conn.Privmsg(channel, user+": "+message)
	_, err := url.Parse(message)
	if err != nil {
		m.log.Debug("not valid url")
		return nil
	}
	return nil
}

func (m *URLTitleModule) Commands() []string {
	return m.commands
}

func (m *URLTitleModule) Event() string {
	return m.event
}

func (m *URLTitleModule) Public() bool {
	return m.public
}
