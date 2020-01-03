package modules

import (
	"time"

	"github.com/huqa/gofibot/internal/pkg/logger"
	"github.com/lrstanley/girc"
)

// ModuleInterface defines a common interface to be used in modules
type ModuleInterface interface {
	Init() error
	Stop() error
	Run(channel, hostmask, user, command, message string, args []string) error
	Event() string
	Commands() []string
	Global() bool
	Schedule() (bool, time.Time)
}

// Module defines a basic struct for modules
type Module struct {
	log      logger.Logger
	commands []string
	client   *girc.Client
	event    string
	global   bool
}
