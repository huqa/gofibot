package gofibot

import (
	"time"

	"github.com/huqa/gofibot/internal/pkg/config"
	"github.com/huqa/gofibot/internal/pkg/logger"
	"github.com/huqa/gofibot/internal/pkg/modules"
	"github.com/lrstanley/girc"
	bolt "go.etcd.io/bbolt"
)

type IRCServiceInterface interface {
	Init() error
	Stop() error
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
	db            *bolt.DB
	callbacks     []string
	location      *time.Location
}

func NewIRCService(log logger.Logger, db *bolt.DB, cfg config.BotConfiguration) IRCServiceInterface {

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

	loc, err := time.LoadLocation(cfg.Location)
	if err != nil {
		log.Error("can't load given timezone", err)
		loc, _ = time.LoadLocation("UTC")
	}

	return &IRCService{
		config:        cfg,
		client:        client,
		callbacks:     make([]string, 0),
		log:           log.Named("ircservice"),
		moduleService: NewModuleService(log, cfg.Channels, cfg.Prefix, loc),
		db:            db,
		location:      loc,
	}
}

func (is *IRCService) Init() error {
	is.log.Info("init bot")

	err := is.LoadModules()
	if err != nil {
		is.log.Error("can't load modules ", err)
		return err
	}
	err = is.JoinChannels()
	if err != nil {
		is.log.Error("error joining channels ", err)
		return err
	}
	return nil
}

func (is *IRCService) Stop() error {
	is.client.Quit("quit")
	is.moduleService.StopModules()
	return nil
}

func (is *IRCService) Connect() error {
	is.log.Info("connecting to server: ", is.config.Server)
	return is.client.Connect()
}

func (is *IRCService) LoadModules() error {
	is.log.Info("loading modules")

	err := is.moduleService.RegisterModules(
		//modules.NewEchoModule(is.log, is.client),
		modules.NewWeatherModule(is.log, is.client),
		modules.NewStatsModule(is.log, is.client, is.db, is.location),
		modules.NewURLTitleModule(is.log, is.client),
		modules.NewDateModule(is.log, is.client, is.location),
		modules.NewGuessModule(is.log, is.client, is.db, is.location),
		modules.NewShouldModule(is.log, is.client),
	)
	if err != nil {
		return err
	}
	is.RegisterModuleCallbacks()
	return nil
}

func (is *IRCService) JoinChannels() error {
	is.log.Info("joining channels when connected: ", is.config.Channels)

	is.client.Handlers.Add(girc.CONNECTED, func(c *girc.Client, e girc.Event) {
		for _, ch := range is.Channels() {
			is.log.Info("joining channel: ", ch)
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

func (is *IRCService) Channels() []string {
	return is.config.Channels
}
