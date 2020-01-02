package gofibot

import (
	"context"
	"database/sql"

	"github.com/huqa/gofibot/internal/pkg/config"

	"github.com/huqa/gofibot/internal/pkg/logger"
)

// Application defines all necessary services to be used by gofibot
type Application struct {
	log        logger.Logger
	ircService IRCServiceInterface
}

// NewApplication construct a new gofibot application
func NewApplication(
	ctx context.Context,
	log logger.Logger,
	db *sql.DB,
	botConfig config.BotConfiguration,
) (a *Application, err error) {
	ircService := NewIRCService(log, botConfig)
	err = ircService.Init()
	if err != nil {
		log.Error("error initializing irc service")
	}

	app := &Application{
		log:        log.Named("gofibot").WithContext(ctx),
		ircService: ircService,
	}

	return app, nil
}
