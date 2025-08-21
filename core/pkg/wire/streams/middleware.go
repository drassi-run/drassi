/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_streams

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/problem"
	"drassi.run/core/pkg/executor/secret"
	"drassi.run/core/pkg/executor/support"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/stream"
)

type Middleware func(handler stream.Handler) stream.Handler

func ProcessCommand(consoleMgr command.ConsoleManager, stack executor.Stack) Middleware {
	return func(handler stream.Handler) stream.Handler {
		return &commandProcessor{
			handler:    handler,
			consoleMgr: consoleMgr,
			stack:      stack,
		}
	}
}

type commandProcessor struct {
	handler    stream.Handler
	consoleMgr command.ConsoleManager
	stack      executor.Stack
}

func (mw *commandProcessor) Handle(ctx context.Context, line string) error {
	cmd := mw.consoleMgr.ParseCommand(line)
	if cmd == nil {
		return mw.handler.Handle(ctx, line)
	}

	if err := mw.consoleMgr.Process(ctx, line, cmd); err != nil {
		if step := mw.stack.Leaf(); step != nil {
			step.SetStatus(records.ResultFailure)
		}
		return err
	}
	return nil
}

func ScanProblem(pm map[string]problem.Matcher, tracker support.Tracker) Middleware {
	return func(handler stream.Handler) stream.Handler {
		return &problemScanner{
			handler: handler,
			matcher: pm,
			tracker: tracker,
		}
	}
}

type problemScanner struct {
	handler stream.Handler
	matcher map[string]problem.Matcher
	tracker support.Tracker
}

func (mw *problemScanner) Handle(ctx context.Context, line string) error {
	err1 := mw.scan(ctx, line)
	err2 := mw.handler.Handle(ctx, line)
	return errors.Join(err1, err2)
}

// https://en.wikipedia.org/wiki/ANSI_escape_code
var colorCodeRegex = regexp.MustCompile(`\033\[[\d;]*m`)

func (mw *problemScanner) scan(ctx context.Context, line string) error {
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
	return mw.tracker.AddIssue(ctx, issue)
}

const skippedIssueMsg = "skipped logging an issue for the matched line because of"

var numberRegex = regexp.MustCompile(`^[+\-]?\d+$`)

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/Handlers/OutputManager.cs#L200
func (mw *problemScanner) toIssuer(pbl *problem.Problem) (*support.Issue, error) {
	if pbl.Message == "" {
		return nil, fmt.Errorf("%s empty message", skippedIssueMsg)
	}
	iss := &support.Issue{
		Message: pbl.Message,
		Data:    make(map[string]string),
	}

	switch strings.ToUpper(pbl.Severity) {
	case "", "ERROR":
		iss.Type = support.IssueTypeError
	case "WARNING":
		iss.Type = support.IssueTypeWarning
	case "NOTICE":
		iss.Type = support.IssueTypeNotice
	default:
		return nil, fmt.Errorf("%s unknown severity %q", skippedIssueMsg, pbl.Severity)
	}

	if !numberRegex.MatchString(pbl.Line) {
		return nil, fmt.Errorf("%s invalid line %q", skippedIssueMsg, pbl.Line)
	} else {
		iss.Data["line"] = pbl.Line
	}
	if !numberRegex.MatchString(pbl.Column) {
		return nil, fmt.Errorf("%s invalid column %q", skippedIssueMsg, pbl.Column)
	} else {
		iss.Data["column"] = pbl.Column
	}
	if code := strings.TrimSpace(pbl.Code); code != "" {
		iss.Data["code"] = code
	}

	iss.Data["file"] = pbl.File // TODO

	return iss, nil
}

func MaskSecret(masker secret.Masker) Middleware {
	return func(handler stream.Handler) stream.Handler {
		return &secretMasker{
			handler: handler,
			masker:  masker,
		}
	}
}

type secretMasker struct {
	handler stream.Handler
	masker  secret.Masker
}

func (mw *secretMasker) Handle(ctx context.Context, line string) error {
	line = mw.masker.Mask(line)
	return mw.handler.Handle(ctx, line)
}
