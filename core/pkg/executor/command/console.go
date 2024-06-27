package command

import (
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"
)

var builtinCommands = []string{
	"stop-commands", "echo",
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Common/ActionCommand.cs#L9
var v1Replacer = strings.NewReplacer(
	"%0D", "\r",
	"%0A", "\n",
	"%5D", "]",
	"%3B", ";",
	"%25", "%",
)

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Common/ActionCommand.cs#L25
var paramReplacer = strings.NewReplacer(
	"%0D", "\r",
	"%0A", "\n",
	"%3A", ":",
	"%2C", ",",
	"%25", "%",
)

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Common/ActionCommand.cs#L18
var valueReplacer = strings.NewReplacer(
	"%0D", "\r",
	"%0A", "\n",
	"%25", "%",
)

type Command struct {
	Name   string
	Params map[string]string
	Value  string
}

type ConsoleCommandHandler = func(cmd *Command) error

type ConsoleCommandManager interface {
	RegisterCommand(name string, echo bool, handler ConsoleCommandHandler) error
	ParseCommand(line string) *Command
	Process(line string, cmd *Command) error
}

func NewConsoleCommandManager(w io.Writer) ConsoleCommandManager {
	mgr := &consoleCommandManager{
		writer:             w,
		registeredCommands: make(map[string]regCmd),
		echo:               true,
		resumeCmdToken:     "",
	}
	_ = mgr.RegisterCommand("stop-commands", true, mgr.stopCommands)
	_ = mgr.RegisterCommand("echo", true, mgr.setEcho)

	return mgr
}

type consoleCommandManager struct {
	writer             io.Writer
	registeredCommands map[string]regCmd

	echo           bool
	resumeCmdToken string
}

type regCmd struct {
	echo    bool
	handler ConsoleCommandHandler
}

func (mgr *consoleCommandManager) RegisterCommand(name string, echo bool, handler ConsoleCommandHandler) error {
	if slices.Contains(builtinCommands, name) {
		if _, ok := mgr.registeredCommands[name]; ok {
			return fmt.Errorf("can't overwrite built-in command %s", name)
		}
	}

	if handler == nil {
		// un-register command
		delete(mgr.registeredCommands, name)
	} else {
		mgr.registeredCommands[name] = regCmd{
			echo:    echo,
			handler: handler,
		}
	}
	return nil
}

func (mgr *consoleCommandManager) ParseCommand(line string) *Command {
	if cmd := mgr.parseCommandV2(line); cmd != nil {
		return cmd
	}
	return mgr.parseCommandV1(line)
}

var cmdV2Regex = regexp.MustCompile(`^::([^\s:]+)( .*)?::(.*)$`)

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Common/ActionCommand.cs#L51
func (mgr *consoleCommandManager) parseCommandV2(line string) *Command {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "::") {
		return nil
	}

	matches := cmdV2Regex.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	name := matches[1]
	if !mgr.isProcessingCommand(name) {
		return nil
	}

	params := mgr.parseCommandParams(matches[2], ",", paramReplacer)
	value := valueReplacer.Replace(matches[3])

	cmd := &Command{
		Name:   name,
		Params: params,
		Value:  value,
	}
	return cmd
}

var cmdV1Regex = regexp.MustCompile(`##\[([^\s\]]+)( .*)?\](.*)$`)

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Common/ActionCommand.cs#L121
func (mgr *consoleCommandManager) parseCommandV1(line string) *Command {
	matches := cmdV1Regex.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	name := matches[1]
	if !mgr.isProcessingCommand(name) {
		return nil
	}

	params := mgr.parseCommandParams(matches[2], ";", v1Replacer)
	value := v1Replacer.Replace(matches[3])

	cmd := &Command{
		Name:   name,
		Params: params,
		Value:  value,
	}
	return cmd
}

func (mgr *consoleCommandManager) isProcessingCommand(cmd string) bool {
	if mgr.resumeCmdToken != "" {
		return cmd == mgr.resumeCmdToken
	}
	_, ok := mgr.registeredCommands[cmd]
	return ok
}

func (mgr *consoleCommandManager) parseCommandParams(params, sep string, replacer *strings.Replacer) map[string]string {
	params = strings.TrimSpace(params)
	if params == "" {
		return nil
	}

	m := make(map[string]string)
	for _, prop := range strings.Split(params, sep) {
		prop = strings.TrimSpace(prop)
		if len(prop) == 0 {
			continue
		}
		pair := strings.SplitN(prop, "=", 2)
		if len(pair) != 2 {
			continue
		}

		key := pair[0]
		value := replacer.Replace(pair[1])
		if value == "" {
			continue
		}

		m[key] = value
	}

	if len(m) == 0 {
		return nil
	}
	return m
}

func (mgr *consoleCommandManager) Process(line string, cmd *Command) error {
	var cmdEcho bool
	var handler func(cmd *Command) error

	if c, ok := mgr.registeredCommands[cmd.Name]; !ok {
		return fmt.Errorf("un-registered command %s", cmd.Name)
	} else {
		cmdEcho = c.echo
		handler = c.handler
	}

	if mgr.echo && cmdEcho {
		if _, err := io.WriteString(mgr.writer, line); err != nil {
			return err
		}
	}

	return handler(cmd)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L765
func (mgr *consoleCommandManager) setEcho(cmd *Command) error {
	mod := strings.TrimSpace(cmd.Value)
	mod = strings.ToUpper(mod)
	switch mod {
	case "ON":
		mgr.echo = true
	case "OFF":
		mgr.echo = false
	default:
		return fmt.Errorf("invalid echo command value. Possible values can be: 'on', 'off'. Got %s", mod)
	}
	return nil
}

func (mgr *consoleCommandManager) stopCommands(cmd *Command) error {
	token := cmd.Value
	if !mgr.validStopCommandToken(token) {
		return fmt.Errorf("invalid stop-command token %s", token)
	}

	mgr.resumeCmdToken = token
	return mgr.RegisterCommand(token, true, mgr.resumeCommands)
}

func (mgr *consoleCommandManager) resumeCommands(cmd *Command) error {
	mgr.resumeCmdToken = ""
	return mgr.RegisterCommand(cmd.Name, true, nil)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L156
func (mgr *consoleCommandManager) validStopCommandToken(token string) bool {
	if token == "" || strings.EqualFold(token, "pause-logging") {
		return false
	}
	_, exists := mgr.registeredCommands[token]
	return !exists
}
