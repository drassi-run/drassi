package wire_streams

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/problem"
	"drassi.run/core/pkg/executor/reporter"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/stream"
)

type Middleware func(handler stream.Handler) stream.Handler

func ProcessCommand(consoleMgr command.ConsoleManager, sup executor.Supervisor) Middleware {
	return func(handler stream.Handler) stream.Handler {
		return &commandProcessor{
			handler:    handler,
			consoleMgr: consoleMgr,
			sup:        sup,
		}
	}
}

type commandProcessor struct {
	handler    stream.Handler
	consoleMgr command.ConsoleManager
	sup        executor.Supervisor
}

func (mw *commandProcessor) Handle(line string) error {
	cmd := mw.consoleMgr.ParseCommand(line)
	if cmd == nil {
		return mw.handler.Handle(line)
	}

	ctx := mw.sup.Context()
	if err := mw.consoleMgr.Process(ctx, line, cmd); err != nil {
		if step := mw.sup.CurrentStep(); step != nil {
			step.SetStatus(records.ResultFailure)
		}
		return err
	}
	return nil
}

func ScanProblem(pm map[string]problem.Matcher, rep reporter.Reporter) Middleware {
	return func(handler stream.Handler) stream.Handler {
		return &problemScanner{
			hdl: handler,
			pm:  pm,
			rep: rep,
		}
	}
}

type problemScanner struct {
	hdl stream.Handler
	pm  map[string]problem.Matcher
	rep reporter.Reporter
}

func (mw *problemScanner) Handle(line string) error {
	err1 := mw.scan(line)
	err2 := mw.hdl.Handle(line)
	return errors.Join(err1, err2)
}

// https://en.wikipedia.org/wiki/ANSI_escape_code
var colorCodeRegex = regexp.MustCompile(`\033\[[\d;]*m`)

func (mw *problemScanner) scan(line string) error {
	var owner string
	var pbl *problem.Problem

	line = colorCodeRegex.ReplaceAllLiteralString(line, "")
	for o, m := range mw.pm {
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
	for o, m := range mw.pm {
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
	return mw.rep.AddIssue(issue)
}

const skippedIssueMsg = "skipped logging an issue for the matched line because of"

var numberRegex = regexp.MustCompile(`^[+\-]?\d+$`)

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/Handlers/OutputManager.cs#L200
func (mw *problemScanner) toIssuer(pbl *problem.Problem) (*reporter.Issue, error) {
	if pbl.Message == "" {
		return nil, fmt.Errorf("%s empty message", skippedIssueMsg)
	}
	iss := &reporter.Issue{
		Message: pbl.Message,
		Data:    make(map[string]string),
	}

	switch strings.ToUpper(pbl.Severity) {
	case "", "ERROR":
		iss.Type = reporter.IssueTypeError
	case "WARNING":
		iss.Type = reporter.IssueTypeWarning
	case "NOTICE":
		iss.Type = reporter.IssueTypeNotice
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
