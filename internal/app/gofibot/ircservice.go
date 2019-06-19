package gofibot

import (
	"github.com/huqa/gofibot/internal/pkg/config"
        "github.com/huqa/gofibot/internal/pkg/modules"
        "github.com/huqa/gofibot/internal/pkg/logger"
	irc "github.com/thoj/go-ircevent"
)

type IRCServiceInterface interface {
	//AddCallback(eventCode string, cb func(e *irc.Event)) int
        AddPRIVMSGModule(module modules.Module) int
	Connect(host string) error
}

type IRCService struct {
        log        logger.Logger
	connection *irc.Connection
	callbacks  []int
        Prefix     string
}

func NewIRCService(log logger.Logger, cfg config.BotConfiguration) IRCServiceInterface {
	conn := irc.IRC(cfg.Nick, cfg.Ident)
	conn.VerboseCallbackHandler = true
	conn.Debug = true
	conn.UseTLS = false

	return &IRCService{
		connection: conn,
		callbacks:  make([]int, 0),
                log: log.Named("ircservice"),
                Prefix: cfg.Prefix,
	}
}

func (is *IRCService) AddCallback(eventCode string, cb func(e *irc.Event)) int {
	cbID := is.Connection.AddCallback(eventCode, func(event *irc.Event) {
		go cb(event)
	})
	is.Callbacks = append(is.Callbacks, cbID)
	return cbID
}

func (is *IRCService) LoadModules() error {
    echoModule := NewEchoModule(is.log, is.connection)
    echoModule.Init()
    _ = is.AddPRIVMSGModule(echoModule)
    return nil
}

func (is *IRCService) Connect(host string) error {
	return is.Connect(host)
}

func (is *IRCService) AddPRIVMSGModule(module modules.Module) int {
        cbID := is.Connection.AddCallback(eventCode, func(event *irc.Event) {
                go func(e *irc.Event) {

                    withoutPrefix := strings.Replace(e.Message(), is.Prefix, "", 1)
                    command := strings.Split(withoutPrefix, " ")[0]

                    if command == module.Command() {
                        err := module.Run(e.Nick, e.Arguments[0], event.Message())
                        if err != nil {
                            log.Error("module run error: ", err)
                        }
                    }
                }(event)
        })
        is.Callbacks = append(is.Callbacks, cbID)
        return cbID
}
