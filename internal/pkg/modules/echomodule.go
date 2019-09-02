package modules

import (
	"github.com/huqa/gofibot/internal/pkg/logger"
	irc "github.com/thoj/go-ircevent"
)

type EchoModule struct {
	log      logger.Logger
	commands []string
	conn     *irc.Connection
	event    string
	public   bool
}

func NewEchoModule(log logger.Logger, conn *irc.Connection) *EchoModule {
	return &EchoModule{
		log:      log.Named("echomodule"),
		commands: []string{"echo"},
		conn:     conn,
		event:    "PRIVMSG",
		public:   false,
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

func (m *EchoModule) Event() string {
	return m.event
}

func (m *EchoModule) Public() bool {
	return m.public
}
