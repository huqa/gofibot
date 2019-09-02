package gofibot

import (
	"github.com/huqa/gofibot/internal/pkg/config"
	"github.com/huqa/gofibot/internal/pkg/logger"
	"github.com/huqa/gofibot/internal/pkg/modules"
	irc "github.com/thoj/go-ircevent"
)

type IRCServiceInterface interface {
	Init() error
	Connect(host string) error
	LoadModules() error
	JoinChannels() error
	RegisterModuleCallbacks()
	AddCallback(eventCode string, cb func(e *irc.Event)) int
}

type IRCService struct {
	log           logger.Logger
	moduleService ModuleServiceInterface
	config        config.BotConfiguration
	connection    *irc.Connection
	callbacks     []int
}

func NewIRCService(log logger.Logger, cfg config.BotConfiguration) IRCServiceInterface {
	conn := irc.IRC(cfg.Nick, cfg.Ident)
	conn.VerboseCallbackHandler = true
	conn.Debug = true
	conn.UseTLS = false

	return &IRCService{
		config:        cfg,
		connection:    conn,
		callbacks:     make([]int, 0),
		log:           log.Named("ircservice"),
		moduleService: NewModuleService(log, cfg.Prefix),
	}
}

func (is *IRCService) Init() error {
	log := is.log.Named("Init")
	log.Info("init bot")

	err := is.LoadModules()
	if err != nil {
		log.Error("can't load modules ", err)
		return err
	}

	err = is.Connect(is.config.Server)
	if err != nil {
		log.Error("error connecting to server ", err)
		return err
	}

	err = is.JoinChannels()
	if err != nil {
		log.Error("error joining channels ", err)
		return err
	}
	return nil
}

func (is *IRCService) AddCallback(eventCode string, cb func(e *irc.Event)) int {
	cbID := is.connection.AddCallback(eventCode, func(event *irc.Event) {
		go cb(event)
	})
	is.callbacks = append(is.callbacks, cbID)
	return cbID
}

func (is *IRCService) LoadModules() error {
	log := is.log.Named("LoadModules")
	log.Info("loading modules")
	err := is.moduleService.RegisterModules(
		modules.NewEchoModule(is.log, is.connection),
		modules.NewWeatherModule(is.log, is.connection),
		modules.NewURLTitleModule(is.log, is.connection),
	)
	if err != nil {
		return err
	}
	is.RegisterModuleCallbacks()
	return nil
}

func (is *IRCService) JoinChannels() error {
	log := is.log.Named("JoinChannels")
	log.Info("joining channels: ", is.config.Channels)

	for _, ch := range is.config.Channels {
		is.connection.Join(ch)
	}
	return nil
}

func (is *IRCService) Connect(host string) error {
	log := is.log.Named("Connect")
	log.Info("connecting to server ", host)
	return is.connection.Connect(host)
}

func (is *IRCService) RegisterModuleCallbacks() {
	cbID := is.connection.AddCallback("PRIVMSG", func(event *irc.Event) {
		go is.moduleService.PRIVMSGCallback(event)
	})
	is.callbacks = append(is.callbacks, cbID)
}
