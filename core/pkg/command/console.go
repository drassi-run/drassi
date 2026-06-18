/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package command

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"drassi.run/core/pkg/scribe"
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

type ConsoleRun[R any] func(ctx context.Context, res R, cmd *Command) error

type ConsoleHandler[R any] struct {
	name string
	echo bool
	run  ConsoleRun[R]
}

func (h *ConsoleHandler[R]) Run(ctx context.Context, res R, cmd *Command) error {
	return h.run(ctx, res, cmd)
}

func NewConsoleHandler[R any](name string, echo bool, run ConsoleRun[R]) *ConsoleHandler[R] {
	return &ConsoleHandler[R]{
		name: name,
		echo: echo,
		run:  run,
	}
}

type ConsoleManager[R any] interface {
	Register(handler *ConsoleHandler[R]) error
	ParseCommand(line string) *Command
	Process(ctx context.Context, res R, line string, cmd *Command) error
}

func NewConsoleManager[R any]() ConsoleManager[R] {
	mgr := &consoleManager[R]{
		registeredCommands: make(map[string]*ConsoleHandler[R]),
		echo:               false, // default to false, unless runner.Debug is set
		resumeCmdToken:     "",
	}
	_ = mgr.Register(NewConsoleHandler("stop-commands", true, mgr.stopCommands))
	_ = mgr.Register(NewConsoleHandler("echo", true, mgr.setEcho))

	return mgr
}

type consoleManager[R any] struct {
	registeredCommands map[string]*ConsoleHandler[R]

	echo           bool
	resumeCmdToken string
}

func (mgr *consoleManager[R]) Register(handler *ConsoleHandler[R]) error {
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

func (mgr *consoleManager[R]) ParseCommand(line string) *Command {
	if cmd := mgr.parseCommandV2(line); cmd != nil {
		return cmd
	}
	return mgr.parseCommandV1(line)
}

var cmdV2Regex = regexp.MustCompile(`^::([^\s:]+)( .*)?::(.*)$`)

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Common/ActionCommand.cs#L51
func (mgr *consoleManager[R]) parseCommandV2(line string) *Command {
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
func (mgr *consoleManager[R]) parseCommandV1(line string) *Command {
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

func (mgr *consoleManager[R]) isProcessingCommand(cmd string) bool {
	if mgr.resumeCmdToken != "" {
		return cmd == mgr.resumeCmdToken
	}
	_, ok := mgr.registeredCommands[cmd]
	return ok
}

func (mgr *consoleManager[R]) parseCommandParams(params, sep string, replacer *strings.Replacer) map[string]string {
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

func (mgr *consoleManager[R]) Process(ctx context.Context, res R, line string, cmd *Command) (err error) {
	ctx, span := xotel.StartSpan(ctx, "ConsoleCommand.Process",
		trace.WithAttributes(xotel.Command(cmd.Name)),
	)
	defer xotel.EndSpan(span, &err)

	handler, ok := mgr.registeredCommands[cmd.Name]
	if !ok {
		return fmt.Errorf("%w %q", ErrNotRegisteredCommand, cmd.Name)
	}

	if mgr.echo && handler.echo {
		scribe.Writef(ctx, "%s", line)
	}

	return handler.run(ctx, res, cmd)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L765
func (mgr *consoleManager[R]) setEcho(_ context.Context, _ R, cmd *Command) error {
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

func (mgr *consoleManager[R]) stopCommands(_ context.Context, _ R, cmd *Command) error {
	token := cmd.Value
	if !mgr.validStopCommandToken(token) {
		return fmt.Errorf("%w %q: invalid token %q", ErrInvalidCommand, "stop", token)
	}

	mgr.resumeCmdToken = token
	return mgr.Register(NewConsoleHandler(token, true, mgr.resumeCommands))
}

func (mgr *consoleManager[R]) resumeCommands(_ context.Context, _ R, cmd *Command) error {
	mgr.resumeCmdToken = ""
	return mgr.Register(NewConsoleHandler[R](cmd.Name, true, nil))
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionCommandManager.cs#L156
func (mgr *consoleManager[R]) validStopCommandToken(token string) bool {
	if token == "" || strings.EqualFold(token, "pause-logging") {
		return false
	}
	_, exists := mgr.registeredCommands[token]
	return !exists
}
