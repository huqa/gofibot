package modules

import (
	"github.com/huqa/gofibot/internal/pkg/logger"
	irc "github.com/thoj/go-ircevent"
)

type EchoModule struct {
	log     logger.Logger
	command string
	conn    *irc.Connection
}

func NewEchoModule(log logger.Logger, conn *irc.Connection) *EchoModule {
	return &EchoModule{
		log:     log.Named("echomodule"),
		command: "echo",
		conn:    conn,
	}
}

func (m *EchoModule) Init() error {
	m.log.Info("Init")
	return nil
}

func (m *EchoModule) Run(user, channel, message string) error {
	m.conn.Privmsg(channel, user+": "+message)
	return nil
}

func (m *EchoModule) Command() string {
	return m.command
}
