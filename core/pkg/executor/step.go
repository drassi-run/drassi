package executor

import (
	"context"
	"strings"

	"drassi.run/core/pkg/executor/evaluator"
	"drassi.run/core/pkg/executor/logging"
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/model/workflows"
	"go.uber.org/dig"
)

type StepRun interface {
	StepId() string
	Base() *BaseStepRun
	DisplayName(stage Stage) string

	Initialize(context.Context, *dig.Scope) error
	PreTask() *Task
	MainTask() *Task
	PostTask() *Task
}

// ensure StepRun implementations
var (
	_ StepRun = (*ActionStepRun)(nil)
	_ StepRun = (*ScriptStepRun)(nil)
	_ StepRun = (*DockerStepRun)(nil)
	_ StepRun = (*NodeStepRun)(nil)
	_ StepRun = (*CompositeStepRun)(nil)
)

type BaseStepRun struct {
	Id               string
	Uid              string
	Name             workflows.Evaluable[string]
	Condition        workflows.Conditional
	ContinueOnError  workflows.Evaluable[bool]
	TimeoutInMinutes workflows.Evaluable[int64]
	Env              workflows.Evaluable[map[string]string]
	Inputs           workflows.Evaluable[map[string]string]
	Outputs          workflows.Evaluable[map[string]string]

	// compute attributes
	displayName string
}

func (s *BaseStepRun) StepId() string {
	return s.Id
}

func (s *BaseStepRun) Base() *BaseStepRun {
	return s
}

func (s *BaseStepRun) DisplayName(stage Stage) string {
	switch stage {
	case StagePre:
		return "Pre " + s.displayName
	case StagePost:
		return "Post " + s.displayName
	default:
		return s.displayName
	}
}

func (s *BaseStepRun) evaluateDisplayName(exprEnv expression.Env, defaultName string, logger logging.Logger) error {
	if s.displayName != "" {
		return nil
	}

	prefix, name := "", ""
	if s.Name == nil {
		prefix, name = "Run ", defaultName
	} else {
		logging.Debugf(logger, "Evaluating display name")
		if err := evaluator.Evaluate(exprEnv, s.Name, &name); err != nil {
			return err
		}
	}

	name = strings.TrimLeft(name, " \t\r\n")
	name, _, _ = strings.Cut(name, "\n")
	name = strings.TrimSpace(name)

	s.displayName = prefix + name
	logging.Debugf(logger, "Set step %q display name to: %q", s.Id, s.displayName)
	return nil
}
