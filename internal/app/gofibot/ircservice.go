package gofibot

import (
	"github.com/huqa/gofibot/internal/pkg/config"
	irc "github.com/thoj/go-ircevent"
)

type IRCServiceInterface interface {
	AddCallback(eventCode string, cb func(e *irc.Event)) int
	Connect(host string) error
}

type IRCService struct {
	Connection *irc.Connection
	Callbacks  []int
}

func NewIRCService(cfg config.BotConfiguration) IRCServiceInterface {
	conn := irc.IRC(cfg.Nick, cfg.Ident)
	conn.VerboseCallbackHandler = true
	conn.Debug = true
	conn.UseTLS = false
	return &IRCService{
		Connection: conn,
		Callbacks:  make([]int, 0),
	}
}

func (is *IRCService) AddCallback(eventCode string, cb func(e *irc.Event)) int {
	cbID := is.Connection.AddCallback(eventCode, func(event *irc.Event) {
		go cb(event)
	})
	is.Callbacks = append(is.Callbacks, cbID)
	return cbID
}

func (is *IRCService) Connect(host string) error {
	return is.Connect(host)
}
