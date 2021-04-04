package modules

import (
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/huqa/gofibot/internal/pkg/logger"
	"github.com/lrstanley/girc"
)

const userAgent string = "Mozilla/5.0 (Linux; Android 7.1.2; DSCS9 Build/NHG47L; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/80.0.3987.149 Safari/537.36"
const setCookie string = "Set-Cookie"
const consentCookie string = "CONSENT=YES; Domain=.youtube.com; Path=/; SameSite=None; Secure; Expires=Sun, 10 Jan 2038 07:59:59 GMT; Max-Age=946080000"

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
		colly.UserAgent(userAgent),
	)
	c.AllowURLRevisit = true
	c.OnHTML("title", m.URLTitleCallback)
	m.titleCollector = c
	y := colly.NewCollector(
		colly.Async(true),
		colly.AllowURLRevisit(),
		colly.MaxDepth(1),
		colly.UserAgent(userAgent),
	)
	y.AllowURLRevisit = true
	y.DisableCookies()
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
			headers := map[string][]string{
				setCookie: {consentCookie},
			}
			m.ytCollector.Request("GET", URL.String(), nil, ctx, headers)
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
