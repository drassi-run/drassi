package command

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"

	"drassi.run/core/util/otel"
	"go.opentelemetry.io/otel/trace"
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

type ConsoleHandler struct {
	name string
	echo bool
	run  func(ctx context.Context, cmd *Command) error
}

func NewConsoleHandler(name string, echo bool, run func(context.Context, *Command) error) *ConsoleHandler {
	return &ConsoleHandler{
		name: name,
		echo: echo,
		run:  run,
	}
}

type ConsoleManager interface {
	Register(handler *ConsoleHandler) error
	ParseCommand(line string) *Command
	Process(ctx context.Context, line string, cmd *Command) error
}

func NewConsoleManager(w io.Writer) ConsoleManager {
	mgr := &consoleManager{
		writer:             w,
		registeredCommands: make(map[string]*ConsoleHandler),
		echo:               false, // default to false, unless runner.Debug is set
		resumeCmdToken:     "",
	}
	_ = mgr.Register(NewConsoleHandler("stop-commands", true, mgr.stopCommands))
	_ = mgr.Register(NewConsoleHandler("echo", true, mgr.setEcho))

	return mgr
}

type consoleManager struct {
	writer             io.Writer
	registeredCommands map[string]*ConsoleHandler

	echo           bool
	resumeCmdToken string
}

func (mgr *consoleManager) Register(handler *ConsoleHandler) error {
	name := handler.name
	if slices.Contains(builtinCommands, name) {
		if _, ok := mgr.registeredCommands[name]; ok {
			return fmt.Errorf("can't overwrite built-in command %q", name)
		}
	}

	if handler.run == nil {
		// un-register command
		delete(mgr.registeredCommands, name)
	} else {
		mgr.registeredCommands[name] = handler
	}
	return nil
}

func (mgr *consoleManager) ParseCommand(line string) *Command {
	if cmd := mgr.parseCommandV2(line); cmd != nil {
		return cmd
	}
	return mgr.parseCommandV1(line)
}

var cmdV2Regex = regexp.MustCompile(`^::([^\s:]+)( .*)?::(.*)$`)

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Common/ActionCommand.cs#L51
func (mgr *consoleManager) parseCommandV2(line string) *Command {
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
func (mgr *consoleManager) parseCommandV1(line string) *Command {
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

func (mgr *consoleManager) isProcessingCommand(cmd string) bool {
	if mgr.resumeCmdToken != "" {
		return cmd == mgr.resumeCmdToken
	}
	_, ok := mgr.registeredCommands[cmd]
	return ok
}

func (mgr *consoleManager) parseCommandParams(params, sep string, replacer *strings.Replacer) map[string]string {
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

func (mgr *consoleManager) Process(ctx context.Context, line string, cmd *Command) (err error) {
	ctx, span := xotel.StartSpan(ctx, "ConsoleCommand.Process",
		trace.WithAttributes(xotel.DrassiCommand(cmd.Name)),
	)
	defer xotel.EndSpan(span, &err)

	handler, ok := mgr.registeredCommands[cmd.Name]
	if !ok {
		return fmt.Errorf("%w %q", ErrNotRegisteredCommand, cmd.Name)
	}

	if mgr.echo && handler.echo {
		if _, err := io.WriteString(mgr.writer, line); err != nil {
			return err
		}
	}

	return handler.run(ctx, cmd)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L765
func (mgr *consoleManager) setEcho(_ context.Context, cmd *Command) error {
	mod := strings.TrimSpace(cmd.Value)
	mod = strings.ToUpper(mod)
	switch mod {
	case "ON":
		mgr.echo = true
	case "OFF":
		mgr.echo = false
	default:
		return fmt.Errorf("%w %q: invalid value %q, should be either 'on' or 'off'", ErrInvalidCommand, "echo", mod)
	}
	return nil
}

func (mgr *consoleManager) stopCommands(_ context.Context, cmd *Command) error {
	token := cmd.Value
	if !mgr.validStopCommandToken(token) {
		return fmt.Errorf("%w %q: invalid token %q", ErrInvalidCommand, "stop", token)
	}

	mgr.resumeCmdToken = token
	return mgr.Register(NewConsoleHandler(token, true, mgr.resumeCommands))
}

func (mgr *consoleManager) resumeCommands(_ context.Context, cmd *Command) error {
	mgr.resumeCmdToken = ""
	return mgr.Register(NewConsoleHandler(cmd.Name, true, nil))
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L156
func (mgr *consoleManager) validStopCommandToken(token string) bool {
	if token == "" || strings.EqualFold(token, "pause-logging") {
		return false
	}
	_, exists := mgr.registeredCommands[token]
	return !exists
}

func ConsoleRun(ctx context.Context, h *ConsoleHandler, cmd *Command) error {
	return h.run(ctx, cmd)
}
