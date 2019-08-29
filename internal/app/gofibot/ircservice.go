package gofibot

import (
	"strings"

	"github.com/huqa/gofibot/internal/pkg/config"
	"github.com/huqa/gofibot/internal/pkg/logger"
	"github.com/huqa/gofibot/internal/pkg/modules"
	irc "github.com/thoj/go-ircevent"
)

type IRCServiceInterface interface {
	//AddCallback(eventCode string, cb func(e *irc.Event)) int
	Init() error
	AddPRIVMSGModule(module modules.Module) int
	Connect(host string) error
	LoadModules() error
	JoinChannels() error
}

type IRCService struct {
	log        logger.Logger
	config     config.BotConfiguration
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
		config:     cfg,
		connection: conn,
		callbacks:  make([]int, 0),
		log:        log.Named("ircservice"),
		Prefix:     cfg.Prefix,
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
	echoModule := modules.NewEchoModule(is.log, is.connection)
	echoModule.Init()
	_ = is.AddPRIVMSGModule(echoModule)
	weatherModule := modules.NewWeatherModule(is.log, is.connection)
	weatherModule.Init()
	_ = is.AddPRIVMSGModule(weatherModule)
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

func (is *IRCService) AddPRIVMSGModule(module modules.Module) int {
	log := is.log.Named("AddPRIVMSGModule")
	cbID := is.connection.AddCallback("PRIVMSG", func(event *irc.Event) {
		go func(e *irc.Event) {

			withoutPrefix := strings.Replace(e.Message(), is.Prefix, "", 1)
			command := strings.Split(withoutPrefix, " ")[0]
			args := strings.Split(event.Message(), " ")[1:]

			for _, cmd := range module.Commands() {
				if command == cmd {
					err := module.Run(e.Nick, e.Arguments[0], event.Message(), args)
					if err != nil {
						log.Error("module run error: ", err)
					}
				}
			}
		}(event)
	})
	is.callbacks = append(is.callbacks, cbID)
	return cbID
}
