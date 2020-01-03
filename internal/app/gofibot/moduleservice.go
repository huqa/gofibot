package gofibot

import (
	"strings"

	"github.com/huqa/gofibot/internal/pkg/logger"
	"github.com/huqa/gofibot/internal/pkg/modules"
	"github.com/lrstanley/girc"
)

// ModuleServiceInterface defines an interface for ModuleService
type ModuleServiceInterface interface {
	RegisterModules(botmodules ...modules.ModuleInterface) error
	Command(string) modules.ModuleInterface
	PRIVMSGCallback(e *girc.Event)
	StopModules() error
}

// ModuleService handles gofibots command based modules
type ModuleService struct {
	log            logger.Logger
	commands       map[string]modules.ModuleInterface
	globalCommands []modules.ModuleInterface
	modules        []modules.ModuleInterface
	callbacks      []int
	Prefix         string
}

// NewModuleService constructs new ModuleService
func NewModuleService(log logger.Logger, prefix string) ModuleServiceInterface {
	return &ModuleService{
		log:            log.Named("moduleservice"),
		globalCommands: make([]modules.ModuleInterface, 0),
		commands:       make(map[string]modules.ModuleInterface, 0),
		Prefix:         prefix,
	}
}

// StopModules stops all registered modules
func (m *ModuleService) StopModules() error {
	m.log.Info("stopping modules")
	for _, module := range m.modules {
		err := module.Stop()
		if err != nil {
			m.log.Error("error stopping module: ", err)
		}
	}
	return nil
}

// RegisterModules registers modules to ModuleService
// First registers a module and then calls its Init method
func (m *ModuleService) RegisterModules(botmodules ...modules.ModuleInterface) error {
	for _, md := range botmodules {
		if md.Global() {
			m.globalCommands = append(m.globalCommands, md)
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
	m.modules = botmodules
	return nil
}

// Command checks if str is found from a map and returns that commands module
func (m *ModuleService) Command(str string) modules.ModuleInterface {
	if md, ok := m.commands[str]; ok {
		return md
	}
	return nil
}

// PRIVMSGCallback calls a modules Run function if event or command matches
func (m *ModuleService) PRIVMSGCallback(e *girc.Event) {
	/*m.log.Debug("event ", e)
	m.log.Debug("event ", e.String())
	m.log.Debug("event", e.Command)
	m.log.Debug("event", e.Params)
	m.log.Debug("source", e.Source.Name)
	m.log.Debug("chan", e.Params[0])
	m.log.Debug("args", e.Params[1:])*/

	channel := e.Params[0]
	message := strings.Join(e.Params[1:], " ")
	if !strings.HasPrefix(message, m.Prefix) {
		for _, pcmd := range m.globalCommands {
			//m.log.Debug(e.Source.Name, channel, message, e.Params[1:])
			err := pcmd.Run(channel, e.Source.String(), e.Source.Name, "", message, e.Params[1:])
			if err != nil {
				m.log.Error("module run error: ", err)
			}
		}
		return
	}

	withoutPrefix := strings.Replace(message, m.Prefix, "", 1)
	command := strings.Split(withoutPrefix, " ")[0]
	msm := m.Command(command)
	if msm != nil {
		if msm.Event() != "PRIVMSG" {
			return
		}
		err := msm.Run(channel, e.Source.String(), e.Source.Name, command, message, e.Params[1:])
		if err != nil {
			m.log.Error("module run error: ", err)
		}
	}
}
