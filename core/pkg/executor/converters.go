package executor

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/executor/problem"
	"drassi.run/core/pkg/executor/reporter"
	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/workflows"
	"github.com/google/uuid"
)

func (e *jobExecutor) toContainerConfig(ctx context.Context, container *workflows.Container) (*container.ContainerConfig, error) {
	return nil, nil
}

const skippedIssueMsg = "skipped logging an issue for the matched line because of"

var numberRegex = regexp.MustCompile(`^[+\-]?\d+$`)

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/Handlers/OutputManager.cs#L200
func (e *jobExecutor) toIssuer(pbl *problem.Problem) (*reporter.Issue, error) {
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

func ToJobRun(jobId string, job *workflows.NormalJob) *JobRun {
	idMap := make(map[string]int)
	stepRuns := make([]StepRun, len(job.Steps))

	for i, step := range job.Steps {
		sr := ToStepRun(step)

		// generate StepId if empty
		if sr.StepId() == "" {
			var id string
			switch s := sr.(type) {
			case *ScriptStepRun:
				id = "run"
			case *DockerStepRun:
				id = normalize(s.Image)
			case *ActionStepRun:
				id = normalize(s.Repo.Repo)
			}

			count := idMap[id] + 1
			idMap[id] = count
			if count > 1 {
				id += "_" + strconv.Itoa(count)
			}

			sr.Base().Id = "__" + id
		}

		stepRuns[i] = sr
	}

	uid, _ := uuid.NewRandom()
	return &JobRun{
		Id:        jobId,
		Uid:       uid.String(),
		Name:      job.Name,
		Container: job.Container,
		Services:  job.Services,
		Env:       job.Env,
		Steps:     stepRuns,
		Outputs:   job.Outputs,
		Defaults:  job.Defaults,
	}
}

func ToStepRun(step workflows.Step) StepRun {
	b := step.Base()
	uid, _ := uuid.NewRandom()
	bsr := &BaseStepRun{
		Id:               b.Id,
		Uid:              uid.String(),
		Name:             b.Name,
		Condition:        b.If,
		ContinueOnError:  b.ContinueOnError,
		TimeoutInMinutes: b.TimeoutInMinutes,
		Env:              b.Env,
	}
	switch s := step.(type) {
	case *workflows.RunStep:
		return toScriptStepRun(s, bsr)
	case *workflows.UsesStep:
		bsr.Inputs = s.With
		if strings.HasPrefix(s.Uses, "docker://") {
			return toDockerStepRun(s, bsr)
		} else {
			return toActionStepRun(s, bsr)
		}
	}
	return nil
}

func toScriptStepRun(s *workflows.RunStep, bsr *BaseStepRun) StepRun {
	return &ScriptStepRun{
		BaseStepRun: *bsr,

		Run:        s.Run,
		Shell:      s.Shell,
		WorkingDir: s.WorkingDir,
	}
}

func toDockerStepRun(s *workflows.UsesStep, bsr *BaseStepRun) StepRun {
	return &DockerStepRun{
		BaseStepRun: *bsr,
		Image:       s.Uses,
	}
}

func toActionStepRun(s *workflows.UsesStep, bsr *BaseStepRun) StepRun {
	repo, _ := model.ParseRepository(s.Uses)
	return &ActionStepRun{
		BaseStepRun: *bsr,
		Repo:        repo,
	}
}

// normalize string by remove all special characters
func normalize(s string) string {
	return strings.Map(normalizeReplacer, s)
}

func normalizeReplacer(r rune) rune {
	if ('0' <= r && r <= '9') || ('A' <= r && r <= 'Z') || ('a' <= r && r <= 'z') {
		return r
	}
	return '_'
}
