package gofibot

import (
	"context"

	"github.com/huqa/gofibot/internal/pkg/config"

	"github.com/huqa/gofibot/internal/pkg/logger"
)

type Application struct {
	log        logger.Logger
	ircService IRCServiceInterface
	//eventHandler
}

func NewApplication(
	ctx context.Context,
	log logger.Logger,
	botConfig config.BotConfiguration,
) (a *Application, err error) {
	ircService := NewIRCService(log, botConfig)

	app := &Application{
		log:        log.Named("gofibot").WithContext(ctx),
		ircService: ircService,
	}
	return app, nil
}
