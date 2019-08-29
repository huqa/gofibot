package modules

import (
	"github.com/huqa/gofibot/internal/pkg/logger"
	irc "github.com/thoj/go-ircevent"
)

type EchoModule struct {
	log      logger.Logger
	commands []string
	conn     *irc.Connection
}

func NewEchoModule(log logger.Logger, conn *irc.Connection) *EchoModule {
	return &EchoModule{
		log:      log.Named("echomodule"),
		commands: []string{"echo"},
		conn:     conn,
	}
}

func (m *EchoModule) Init() error {
	m.log.Info("Init")
	return nil
}

func (m *EchoModule) Run(user, channel, message string, args []string) error {
	m.conn.Privmsg(channel, user+": "+message)
	return nil
}

func (m *EchoModule) Commands() []string {
	return m.commands
}
