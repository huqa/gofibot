package gofibot

import (
	"strings"
	"time"

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
	channels       []string
	commands       map[string]modules.ModuleInterface
	globalCommands []modules.ModuleInterface
	modules        []modules.ModuleInterface
	callbacks      []int
	tickers        []*time.Ticker
	timers         []*time.Timer
	Prefix         string
}

// NewModuleService constructs new ModuleService
func NewModuleService(log logger.Logger, channels []string, prefix string) ModuleServiceInterface {
	return &ModuleService{
		log:            log.Named("moduleservice"),
		channels:       channels,
		globalCommands: make([]modules.ModuleInterface, 0),
		commands:       make(map[string]modules.ModuleInterface, 0),
		tickers:        make([]*time.Ticker, 0),
		timers:         make([]*time.Timer, 0),
		Prefix:         prefix,
	}
}

// StopModules stops all registered modules
func (m *ModuleService) StopModules() error {
	m.log.Info("stopping tickers")
	for _, ticker := range m.tickers {
		ticker.Stop()
	}
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
			m.log.Infof("registered public command")
		}
		for _, cmd := range md.Commands() {
			m.commands[cmd] = md
			m.log.Infof("registered command %s", cmd)
		}
		err := md.Init()
		if err != nil {
			return err
		}
		hasSchedule, nextRunTime, duration := md.Schedule()
		if hasSchedule {
			now := time.Now()
			nextRunDuration := nextRunTime.Sub(now)
			m.log.Info("scheduling module to run at ", now.Add(nextRunDuration))
			timer := time.AfterFunc(nextRunDuration, func() { m.schedulePRIVMSG(md, duration) })
			m.timers = append(m.timers, timer)
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

	channel := e.Params[0]
	if !strings.HasPrefix(e.Params[1], m.Prefix) {
		for _, pcmd := range m.globalCommands {
			//m.log.Debug(e.Source.Name, channel, message, e.Params[1:])
			err := pcmd.Run(channel, e.Source.String(), e.Source.Name, "", e.Params[1:])
			if err != nil {
				m.log.Error("module run error: ", err)
			}
		}
		return
	}

	withoutPrefix := strings.Replace(e.Params[1], m.Prefix, "", 1)
	command := strings.Split(withoutPrefix, " ")[0]
	msm := m.Command(command)
	if msm != nil {
		if msm.Event() != "PRIVMSG" {
			return
		}
		err := msm.Run(channel, e.Source.String(), e.Source.Name, command, e.Params[1:])
		if err != nil {
			m.log.Error("module run error: ", err)
		}
	}
}

func (m *ModuleService) schedulePRIVMSG(module modules.ModuleInterface, duration time.Duration) {
	ticker := time.NewTicker(duration)
	m.tickers = append(m.tickers, ticker)
	for ; true; <-ticker.C {
		for _, channel := range m.channels {
			module.Run(channel, "", "", module.Commands()[0], make([]string, 0))
		}
	}
}
