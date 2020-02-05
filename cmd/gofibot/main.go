// Defines the executable for gofibot
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"

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

	db, err := sql.Open("sqlite3", fmt.Sprintf("./db/%s", botConfig.DatabaseFile))
	if err != nil {
		log.Fatal("failed to open sqlite database ", err)
		os.Exit(1)
	}
	defer db.Close()

	_, err = gofibot.NewApplication(ctx, log, db, appConfig.BotConfig)
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
