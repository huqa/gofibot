// Defines the executable for gofibot
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	bolt "go.etcd.io/bbolt"

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

	db, err := bolt.Open(fmt.Sprintf("./db/%s", botConfig.DatabaseFile), 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal("failed to open database ", err)
		os.Exit(1)
	}
	defer db.Close()

	app, err := gofibot.NewApplication(ctx, log, db, appConfig.BotConfig)
	if err != nil {
		log.Fatal("failed to create gofibot ", err)
		os.Exit(1)
	}
	defer app.Shutdown()
	log.Info("started gofibot")

	go func() {
		for {
			if err := app.IRCService.Connect(); err != nil {
				log.Error("connection error", err)
				log.Info("reconnecting in 15 seconds")
				time.Sleep(15 * time.Second)
			}
		}
	}()

	proc.WaitForInterrupt()

	defer log.Info("stopping gofibot")
}
