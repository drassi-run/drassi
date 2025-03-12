package executor

import (
	"context"
	"strings"

	"drassi.run/core/pkg/executor/evaluator"
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/scribe"
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

func (sr *BaseStepRun) StepId() string {
	return sr.Id
}

func (sr *BaseStepRun) Base() *BaseStepRun {
	return sr
}

func (sr *BaseStepRun) DisplayName(stage Stage) string {
	switch stage {
	case StagePre:
		return "Pre " + sr.displayName
	case StagePost:
		return "Post " + sr.displayName
	default:
		return sr.displayName
	}
}

func (sr *BaseStepRun) evaluateDisplayName(ctx context.Context, exprEnv expression.Env, defaultName string) error {
	s := scribe.FromContext(ctx)
	if sr.displayName != "" {
		return nil
	}

	prefix, name := "", ""
	if sr.Name == nil {
		prefix, name = "Run ", defaultName
	} else {
		s.Debugf("Evaluating display name")
		if err := evaluator.Evaluate(exprEnv, sr.Name, &name); err != nil {
			return err
		}
	}

	name = strings.TrimLeft(name, " \t\r\n")
	name, _, _ = strings.Cut(name, "\n")
	name = strings.TrimSpace(name)

	sr.displayName = prefix + name
	s.Debugf("Set step %q display name to: %q", sr.Id, sr.displayName)
	return nil
}
