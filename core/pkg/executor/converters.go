package executor

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/dungdm93/drassi/core/pkg/container"
	"github.com/dungdm93/drassi/core/pkg/executor/problem"
	"github.com/dungdm93/drassi/core/pkg/executor/reporter"
	"github.com/dungdm93/drassi/core/pkg/model/workflows"
)

func (e *JobExecutor) toContainerConfig(ctx context.Context, container *workflows.Container) (*container.ContainerConfig, error) {
	return nil, nil
}

const skippedIssueMsg = "skipped logging an issue for the matched line because of"

var numberRegex = regexp.MustCompile(`^[+\-]?\d+$`)

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/Handlers/OutputManager.cs#L200
func (e *JobExecutor) toIssuer(pbl *problem.Problem) (*reporter.Issue, error) {
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
		return nil, fmt.Errorf("%s unknown severity '%s'", skippedIssueMsg, pbl.Severity)
	}

	if !numberRegex.MatchString(pbl.Line) {
		return nil, fmt.Errorf("%s invalid line '%s'", skippedIssueMsg, pbl.Line)
	} else {
		iss.Data["line"] = pbl.Line
	}
	if !numberRegex.MatchString(pbl.Column) {
		return nil, fmt.Errorf("%s invalid column '%s'", skippedIssueMsg, pbl.Column)
	} else {
		iss.Data["column"] = pbl.Column
	}
	if code := strings.TrimSpace(pbl.Code); code != "" {
		iss.Data["code"] = code
	}

	iss.Data["file"] = pbl.File // TODO

	return iss, nil
}
