package gofibot

import (
	"database/sql"

	"github.com/huqa/gofibot/internal/pkg/config"
	"github.com/huqa/gofibot/internal/pkg/logger"
	"github.com/huqa/gofibot/internal/pkg/modules"
	"github.com/lrstanley/girc"
)

type IRCServiceInterface interface {
	Init() error
	Connect() error
	LoadModules() error
	JoinChannels() error
	RegisterModuleCallbacks()
}

type IRCService struct {
	log           logger.Logger
	moduleService ModuleServiceInterface
	config        config.BotConfiguration
	client        *girc.Client
	db            *sql.DB
	callbacks     []string
}

func NewIRCService(log logger.Logger, db *sql.DB, cfg config.BotConfiguration) IRCServiceInterface {

	config := girc.Config{
		Nick:   cfg.Nick,
		User:   cfg.Ident,
		Name:   cfg.Realname,
		Server: cfg.Server,
		Port:   6667,
		Out:    log,
	}

	log.Debug(config)

	client := girc.New(config)

	return &IRCService{
		config:        cfg,
		client:        client,
		callbacks:     make([]string, 0),
		log:           log.Named("ircservice"),
		moduleService: NewModuleService(log, cfg.Prefix),
		db:            db,
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
	err = is.JoinChannels()
	if err != nil {
		log.Error("error joining channels ", err)
		return err
	}

	err = is.Connect()
	if err != nil {
		log.Error("error connecting to server ", err)
		return err
	}

	return nil
}

func (is *IRCService) Connect() error {
	log := is.log.Named("Connect")
	log.Info("connecting to server: ", is.config.Server)
	return is.client.Connect()
}

func (is *IRCService) LoadModules() error {
	log := is.log.Named("LoadModules")
	log.Info("loading modules")
	err := is.moduleService.RegisterModules(
		modules.NewEchoModule(is.log, is.client),
		modules.NewWeatherModule(is.log, is.client),
		modules.NewURLTitleModule(is.log, is.client),
	)
	if err != nil {
		return err
	}
	is.RegisterModuleCallbacks()
	return nil
}

func (is *IRCService) JoinChannels() error {
	log := is.log.Named("JoinChannels")
	log.Info("joining channels when connected: ", is.config.Channels)

	is.client.Handlers.Add(girc.CONNECTED, func(c *girc.Client, e girc.Event) {
		for _, ch := range is.config.Channels {
			log.Info("joining channel: ", ch)
			c.Cmd.Join(ch)
		}
	})
	return nil
}

func (is *IRCService) RegisterModuleCallbacks() {
	cbID := is.client.Handlers.Add(girc.PRIVMSG, func(c *girc.Client, e girc.Event) {
		go is.moduleService.PRIVMSGCallback(&e)
	})
	is.callbacks = append(is.callbacks, cbID)
}
