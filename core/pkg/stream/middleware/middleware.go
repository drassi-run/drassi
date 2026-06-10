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
	"regexp"
	"strings"

	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/command/issue"
	"drassi.run/core/pkg/executor/problem"
	"drassi.run/core/pkg/executor/secret"
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

func ScanProblem[R any](pm map[string]problem.Matcher, reporter issue.Reporter) Middleware[R] {
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
	matcher  map[string]problem.Matcher
	reporter issue.Reporter
}

func (mw *problemScanner[R]) RHandle(ctx context.Context, res R, line string) error {
	err1 := mw.scan(ctx, line)
	err2 := mw.handler.RHandle(ctx, res, line)
	return errors.Join(err1, err2)
}

// https://en.wikipedia.org/wiki/ANSI_escape_code
var colorCodeRegex = regexp.MustCompile(`\033\[[\d;]*m`)

func (mw *problemScanner[R]) scan(ctx context.Context, line string) error {
	var owner string
	var pbl *problem.Problem

	line = colorCodeRegex.ReplaceAllLiteralString(line, "")
	for o, m := range mw.matcher {
		if p := m.Match(line); p != nil {
			owner, pbl = o, p
			break
		}
	}

	// Not matched
	if pbl == nil {
		return nil
	}

	// Matched
	// 1. Reset other matchers
	for o, m := range mw.matcher {
		if o != owner {
			m.Reset()
		}
	}

	// 2. convert Problem to Issue
	issue, err := mw.toIssuer(pbl)
	if err != nil {
		return err
	}

	// 3. Report the issue
	return mw.reporter.AddIssue(ctx, issue)
}

const skippedIssueMsg = "skipped logging an issue for the matched line because of"

var numberRegex = regexp.MustCompile(`^[+\-]?\d+$`)

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/Handlers/OutputManager.cs#L200
func (mw *problemScanner[R]) toIssuer(pbl *problem.Problem) (*issue.Issue, error) {
	if pbl.Message == "" {
		return nil, fmt.Errorf("%s empty message", skippedIssueMsg)
	}
	iss := &issue.Issue{
		Message: pbl.Message,
		Data:    make(map[string]string),
	}

	switch strings.ToUpper(pbl.Severity) {
	case "", "ERROR":
		iss.Type = issue.TypeError
	case "WARNING":
		iss.Type = issue.TypeWarning
	case "NOTICE":
		iss.Type = issue.TypeNotice
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

	iss.Data["file"] = pbl.File // TODO

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
