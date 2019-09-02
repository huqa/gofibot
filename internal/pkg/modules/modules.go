package modules

type ModuleInterface interface {
	Init() error
	Run(user, channel, message string, args []string) error
	Event() string
	Commands() []string
	Public() bool
}
