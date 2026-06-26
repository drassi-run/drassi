/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package middleware

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"drassi.run/core/pkg/command"
	"drassi.run/core/pkg/command/cmdtypes"
	"drassi.run/core/pkg/problem"
	"drassi.run/core/pkg/secret"
	"drassi.run/core/pkg/stream"
)

type Middleware[R any] func(handler stream.ResourceHandler[R]) stream.ResourceHandler[R]

func ProcessCommand[R any](consMgr command.ConsoleManager[R]) Middleware[R] {
	return func(handler stream.ResourceHandler[R]) stream.ResourceHandler[R] {
		return &commandProcessor[R]{
			handler: handler,
			consMgr: consMgr,
		}
	}
}

type commandProcessor[R any] struct {
	handler stream.ResourceHandler[R]
	consMgr command.ConsoleManager[R]
}

func (mw *commandProcessor[R]) RHandle(ctx context.Context, res R, line string) error {
	cmd := mw.consMgr.ParseCommand(line)
	if cmd == nil {
		return mw.handler.RHandle(ctx, res, line)
	}

	if err := mw.consMgr.Process(ctx, res, line, cmd); err != nil {
		return fmt.Errorf("process command %q: %w", cmd.Name, err)
	}
	return nil
}

func ScanProblem[R any](pm problem.Matchers, reporter cmdtypes.Reporter[R]) Middleware[R] {
	return func(handler stream.ResourceHandler[R]) stream.ResourceHandler[R] {
		return &problemScanner[R]{
			handler:  handler,
			matcher:  pm,
			reporter: reporter,
		}
	}
}

type problemScanner[R any] struct {
	handler  stream.ResourceHandler[R]
	matcher  problem.Matchers
	reporter cmdtypes.Reporter[R]
}

func (mw *problemScanner[R]) RHandle(ctx context.Context, res R, line string) (err error) {
	pbl := mw.scan(line)
	if pbl != nil {
		err = mw.report(ctx, res, pbl)
	}

	err2 := mw.handler.RHandle(ctx, res, line)
	return errors.Join(err, err2)
}

// https://en.wikipedia.org/wiki/ANSI_escape_code
var colorCodeRegex = regexp.MustCompile(`\033\[[\d;]*m`)

func (mw *problemScanner[R]) scan(line string) (pbl *problem.Problem) {
	var owner string

	line = colorCodeRegex.ReplaceAllLiteralString(line, "")
	for o, m := range mw.matcher {
		if p := m.Match(line); p != nil {
			owner, pbl = o, p
			break
		}
	}

	// Not matched
	if pbl == nil {
		return
	}

	// Matched - then reset other matchers
	for o, m := range mw.matcher {
		if o != owner {
			m.Reset()
		}
	}

	return
}

func (mw *problemScanner[R]) report(ctx context.Context, res R, pbl *problem.Problem) error {
	iss, err := mw.toIssuer(pbl)
	if err != nil {
		return err
	}

	return mw.reporter.AddIssue(ctx, res, iss)
}

const skippedIssueMsg = "skipped logging an issue for the matched line because of"

var numberRegex = regexp.MustCompile(`^[+\-]?\d+$`)

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/Handlers/OutputManager.cs#L200
func (mw *problemScanner[R]) toIssuer(pbl *problem.Problem) (*cmdtypes.Issue, error) {
	if pbl.Message == "" {
		return nil, fmt.Errorf("%s empty message", skippedIssueMsg)
	}
	iss := &cmdtypes.Issue{
		Message: pbl.Message,
		Data:    make(map[string]string),
	}

	switch strings.ToUpper(pbl.Severity) {
	case "", "ERROR":
		iss.Type = cmdtypes.IssueTypeError
	case "WARNING":
		iss.Type = cmdtypes.IssueTypeWarning
	case "NOTICE":
		iss.Type = cmdtypes.IssueTypeNotice
	default:
		return nil, fmt.Errorf("%s unknown severity %q", skippedIssueMsg, pbl.Severity)
	}

	if !numberRegex.MatchString(pbl.Line) {
		return nil, fmt.Errorf("%s invalid line %q", skippedIssueMsg, pbl.Line)
	}
	iss.Data["line"] = pbl.Line

	if !numberRegex.MatchString(pbl.Column) {
		return nil, fmt.Errorf("%s invalid column %q", skippedIssueMsg, pbl.Column)
	}
	iss.Data["column"] = pbl.Column

	if code := strings.TrimSpace(pbl.Code); code != "" {
		iss.Data["code"] = code
	}

	file := pbl.File
	if file != "" {
		if pbl.FromPath != "" && !filepath.IsAbs(file) {
			file = filepath.Join(pbl.FromPath, file)
		}
		// NOTE: "file" will be refined later by [wire_command.refinePathReporter]
		iss.Data["file"] = file
	}

	return iss, nil
}

func MaskSecret[R any](masker secret.Masker) Middleware[R] {
	return func(handler stream.ResourceHandler[R]) stream.ResourceHandler[R] {
		return &secretMasker[R]{
			handler: handler,
			masker:  masker,
		}
	}
}

type secretMasker[R any] struct {
	handler stream.ResourceHandler[R]
	masker  secret.Masker
}

func (mw *secretMasker[R]) RHandle(ctx context.Context, res R, line string) error {
	line = mw.masker.Mask(line)
	return mw.handler.RHandle(ctx, res, line)
}
