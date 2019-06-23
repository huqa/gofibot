package modules

type Module interface {
	Init() error
	Run(user, channel, message string) error
	Command() string
}
