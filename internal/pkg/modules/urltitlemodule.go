package modules

import (
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/huqa/gofibot/internal/pkg/logger"
	"github.com/lrstanley/girc"
)

const UserAgent string = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.3177.89 Safari/537.36 Spreaker/1.0"

// URLTitleModule handles url titles scraped from PRIVMSGs
// Todo cache
type URLTitleModule struct {
	*Module
	titleCollector *colly.Collector
	ytCollector    *colly.Collector
}

// URLTitle defines a URLs content
type URLTitle string

var youtubeURLs = map[string]string{
	"www.youtube.com": "",
	"youtube.com":     "",
	"youtu.be":        "",
	"m.youtube.com":   "",
}

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
		colly.UserAgent(UserAgent),
	)
	c.AllowURLRevisit = true
	c.OnHTML("title", m.URLTitleCallback)
	m.titleCollector = c
	y := colly.NewCollector(
		colly.Async(true),
		colly.AllowURLRevisit(),
		colly.MaxDepth(1),
		colly.UserAgent(UserAgent),
	)
	y.AllowURLRevisit = true
	y.OnHTML("meta[name=title]", m.YTTitleCallback)
	m.ytCollector = y
	//c.OnError(func(r *colly.Response, err error) {
	//m.log.Error("error: ", r.StatusCode, err)
	//})
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
		// youtube urls have their own collector
		if _, ok := youtubeURLs[URL.Hostname()]; ok {
			m.ytCollector.Request("GET", URL.String(), nil, ctx, nil)
			continue
		}

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
		title := URLTitle(strings.TrimSpace(e.Text))
		m.client.Cmd.Message(channel, "Title: "+string(title))
	}
	return
}

func (m *URLTitleModule) YTTitleCallback(e *colly.HTMLElement) {
	if e.Index == 0 {
		channel := e.Response.Ctx.Get("Channel")
		title := URLTitle(strings.TrimSpace(e.Attr("content")) + " - YouTube")
		m.client.Cmd.Message(channel, "Title: "+string(title))
	}
	return
}
