package modules

import "time"

// ModuleInterface defines a common interface to be used in modules
type ModuleInterface interface {
	Init() error
	Run(user, channel, message string, args []string) error
	Event() string
	Commands() []string
	Global() bool
	Schedule() (bool, time.Time)
}
