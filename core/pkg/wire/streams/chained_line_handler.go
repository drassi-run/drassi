package wire_streams

import (
	"fmt"
	"regexp"
	"strings"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/problem"
	"drassi.run/core/pkg/executor/reporter"
	"drassi.run/core/pkg/model/records"
)

type chainedLineHandler func(line string) (next bool, err error)

func processCommand(consoleMgr command.ConsoleManager, sup executor.Supervisor) chainedLineHandler {
	return func(line string) (bool, error) {
		cmd := consoleMgr.ParseCommand(line)
		if cmd == nil {
			return true, nil
		}

		err := consoleMgr.Process(line, cmd)
		if err != nil {
			step := sup.CurrentStep()
			if step != nil {
				step.SetStatus(records.ResultFailure)
			}
		}
		return false, err
	}
}

// https://en.wikipedia.org/wiki/ANSI_escape_code
var colorCodeRegex = regexp.MustCompile(`\033\[[\d;]*m`)

func scanProblem(pm map[string]problem.Matcher, rep reporter.Reporter) chainedLineHandler {
	return func(line string) (bool, error) {
		line = colorCodeRegex.ReplaceAllLiteralString(line, "")
		var owner string
		var pbl *problem.Problem

		for o, m := range pm {
			if p := m.Match(line); p != nil {
				owner = o
				pbl = p
				break
			}
		}

		// Not matched
		if pbl == nil {
			return true, nil
		}

		// Matched
		// 1. Reset other matchers
		for o, m := range pm {
			if o != owner {
				m.Reset()
			}
		}

		// 2. convert Problem to Issue
		issue, err := toIssuer(pbl)
		if err != nil {
			return true, err
		}

		// 3. Report the issue
		err = rep.AddIssue(issue)
		return true, err
	}
}

const skippedIssueMsg = "skipped logging an issue for the matched line because of"

var numberRegex = regexp.MustCompile(`^[+\-]?\d+$`)

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/Handlers/OutputManager.cs#L200
func toIssuer(pbl *problem.Problem) (*reporter.Issue, error) {
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
