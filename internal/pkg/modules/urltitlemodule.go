package modules

import (
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/huqa/gofibot/internal/pkg/logger"
	"github.com/lrstanley/girc"
)

// URLTitleModule handles url titles scraped from PRIVMSGs
// Todo cache
type URLTitleModule struct {
	*Module
	titleCollector *colly.Collector
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
	}
}

// Init initializes url title module
func (m *URLTitleModule) Init() error {
	m.log.Info("Init")
	c := colly.NewCollector(
		colly.Async(true),
		colly.AllowURLRevisit(),
		colly.MaxDepth(1),
	)
	c.AllowURLRevisit = true
	c.OnHTML("title", m.URLTitleCallback)
	//c.OnError(func(r *colly.Response, err error) {
	//m.log.Error("error: ", r.StatusCode, err)
	//})
	m.titleCollector = c
	return nil
}

// Stop is run when module is stopped
func (m *URLTitleModule) Stop() error {
	return nil
}

// Run shouts url titles to PRIVMSG in target channel
// TODO: imdb url support
func (m *URLTitleModule) Run(channel, hostmask, user, command string, args []string) error {
	//message := strings.Join(args, " ")
	for _, message := range args {
		URL, err := url.Parse(message)
		if err != nil {
			continue
		}
		ctx := colly.NewContext()
		ctx.Put("Channel", channel)

		m.titleCollector.Request("GET", URL.String(), nil, ctx, nil)
	}
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

func (m *URLTitleModule) URLTitleCallback(e *colly.HTMLElement) {
	if e.Index == 0 {
		channel := e.Response.Ctx.Get("Channel")
		title := URLTitle(strings.TrimLeft(strings.TrimLeft(e.Text, " "), "\t"))
		m.client.Cmd.Message(channel, "Title: "+string(title))
	}
	return
}
