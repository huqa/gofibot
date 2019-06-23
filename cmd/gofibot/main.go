// Defines the executable for gofibot
package main

import (
	"context"
	"flag"
	"os"

	gofibot "github.com/huqa/gofibot/internal/app/gofibot"
	"github.com/huqa/gofibot/internal/pkg/config"
	"github.com/huqa/gofibot/internal/pkg/logger"
	"github.com/huqa/gofibot/internal/pkg/proc"
)

type configuration struct {
	Logger    logger.Configuration
	BotConfig config.BotConfiguration
}

var defaultConfiguration = configuration{
	Logger: logger.DefaultConfiguration(),
}

func main() {
	ctx := context.Background()
	appConfig := defaultConfiguration

	// read command line flags
	botConfigFilePath := flag.String(
		"bot-config",
		"./config/bot-config.json",
		"bot config file path")
	flag.Parse()

	botConfig, err := config.LoadBotConfiguration(*botConfigFilePath)
	if err != nil {
		logger.Fatal("can't load bot configuration: ", err)
		os.Exit(1)
	}

	appConfig.BotConfig = botConfig

	log := logger.New(appConfig.Logger)
	_, err = gofibot.NewApplication(ctx, log, appConfig.BotConfig)
	if err != nil {
		log.Fatal("failed to create gofibot ", err)
		os.Exit(1)
	}
	log.Info("started gofibot")

	// create bot
	proc.WaitForInterrupt()

	log.Info("stopping gofibot")
	defer log.Info("stopped gofibot")
}
