package gofibot

import (
	"context"

	"github.com/huqa/gofibot/internal/pkg/config"
	bolt "go.etcd.io/bbolt"

	"github.com/huqa/gofibot/internal/pkg/logger"
)

// Application defines all necessary services to be used by gofibot
type Application struct {
	log        logger.Logger
	IRCService IRCServiceInterface
}

// NewApplication construct a new gofibot application
func NewApplication(
	ctx context.Context,
	log logger.Logger,
	db *bolt.DB,
	botConfig config.BotConfiguration,
) (app *Application, err error) {
	ircService := NewIRCService(log, db, botConfig)
	err = ircService.Init()
	if err != nil {
		log.Error("error initializing irc service")
		return app, err
	}

	app = &Application{
		log:        log.Named("gofibot").WithContext(ctx),
		IRCService: ircService,
	}

	return app, nil
}

func (a *Application) Shutdown() {
	a.IRCService.Stop()
}
