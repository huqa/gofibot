package modules

type Module interface {
	Init() error
	Run(user, channel, message string, args []string) error
	Commands() []string
}
