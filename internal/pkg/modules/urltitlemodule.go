package modules

import (
	"net/url"
	"strconv"
	"time"

	"github.com/gocolly/colly"
	"github.com/huqa/gofibot/internal/pkg/logger"
	"github.com/lrstanley/girc"
)

// URLTitleModule handles url titles scraped from PRIVMSGs
// Todo cache
type URLTitleModule struct {
	*Module
	titleCollector *colly.Collector
	responses      map[string]URLTitle
}

// URLTitle defines a URLs content
type URLTitle string

// NewURLTitleModule constructs new URLTitleModule
func NewURLTitleModule(log logger.Logger, client *girc.Client) *URLTitleModule {
	return &URLTitleModule{
		&Module{
			log:    log.Named("urltitlemodule"),
			client: client,
			global: true,
			event:  "PRIVMSG",
		},
		nil,
		make(map[string]URLTitle, 0),
	}
}

// Init initializes url title module
func (m *URLTitleModule) Init() error {
	m.log.Info("Init")
	c := colly.NewCollector(
		colly.Async(true),
		colly.AllowURLRevisit(),
	)
	c.AllowURLRevisit = true
	c.OnHTML("title", func(e *colly.HTMLElement) {
		var ID = e.Response.Ctx.Get("ID")
		var Channel = e.Response.Ctx.Get("Channel")
		m.responses[ID+Channel] = URLTitle(e.Text)
	})
	m.titleCollector = c
	return nil
}

// Stop is run when module is stopped
func (m *URLTitleModule) Stop() error {
	return nil
}

// Run shouts url titles to PRIVMSG in target channel
// TODO: imdb url support
func (m *URLTitleModule) Run(channel, hostmask, user, command, message string, args []string) error {
	//m.conn.Privmsg(channel, user+": "+message)
	URL, err := url.Parse(message)
	if err != nil {
		m.log.Debug("checked privmsg for urltitle - not found")
		return nil
	}
	ID := strconv.FormatInt(time.Now().UnixNano(), 10)
	ctx := colly.NewContext()
	ctx.Put("ID", ID)
	ctx.Put("Channel", channel)
	key := ID + channel
	m.titleCollector.Request("GET", URL.String(), nil, ctx, nil)
	m.titleCollector.Wait()
	URLTitle, ok := m.responses[key]
	if !ok {
		return nil
	}
	delete(m.responses, key)
	m.client.Cmd.Message(channel, "Title: "+string(URLTitle))
	return nil
}

// Commands returns all commands used by this module
func (m *URLTitleModule) Commands() []string {
	return m.commands
}

// Event returns event type used by this module
func (m *URLTitleModule) Event() string {
	return m.event
}

// Global returns true if this module is a global command
func (m *URLTitleModule) Global() bool {
	return m.global
}

func (m *URLTitleModule) Schedule() (bool, time.Time, time.Duration) {
	return false, time.Time{}, 0
}
