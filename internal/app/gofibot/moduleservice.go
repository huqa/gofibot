package gofibot

import (
	"strings"

	"github.com/huqa/gofibot/internal/pkg/logger"
	"github.com/huqa/gofibot/internal/pkg/modules"
	irc "github.com/thoj/go-ircevent"
)

type ModuleServiceInterface interface {
	RegisterModules(botmodules ...modules.ModuleInterface) error
	Command(string) modules.ModuleInterface
	PRIVMSGCallback(e *irc.Event)
}

type ModuleService struct {
	log            logger.Logger
	commands       map[string]modules.ModuleInterface
	publicCommands []modules.ModuleInterface
	callbacks      []int
	Prefix         string
}

func NewModuleService(log logger.Logger, prefix string) ModuleServiceInterface {
	return &ModuleService{
		log:            log.Named("moduleservice"),
		publicCommands: make([]modules.ModuleInterface, 0),
		commands:       make(map[string]modules.ModuleInterface, 0),
		Prefix:         prefix,
	}
}

func (m *ModuleService) RegisterModules(botmodules ...modules.ModuleInterface) error {
	for _, md := range botmodules {
		if md.Public() {
			m.publicCommands = append(m.publicCommands, md)
			m.log.Infof("registered public command %s", md)
		}
		for _, cmd := range md.Commands() {
			m.commands[cmd] = md
			m.log.Infof("registered command %s", cmd)
		}
		err := md.Init()
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *ModuleService) Command(str string) modules.ModuleInterface {
	if md, ok := m.commands[str]; ok {
		return md
	}
	return nil
}

func (m *ModuleService) PRIVMSGCallback(e *irc.Event) {
	if !strings.HasPrefix(e.Message(), m.Prefix) {
		args := strings.Split(e.Message(), " ")[1:]
		for _, pcmd := range m.publicCommands {
			err := pcmd.Run(e.Nick, e.Arguments[0], e.Message(), args)
			if err != nil {
				m.log.Error("module run error: ", err)
			}
		}
		return
	}

	withoutPrefix := strings.Replace(e.Message(), m.Prefix, "", 1)
	command := strings.Split(withoutPrefix, " ")[0]
	args := strings.Split(e.Message(), " ")[1:]

	msm := m.Command(command)
	if msm != nil {
		if msm.Event() != "PRIVMSG" {
			return
		}
		err := msm.Run(e.Nick, e.Arguments[0], e.Message(), args)
		if err != nil {
			m.log.Error("module run error: ", err)
		}
	}
}
