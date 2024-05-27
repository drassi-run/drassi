package executor

import (
	"context"
	"fmt"
	"strings"

	"github.com/dungdm93/drassi/core/pkg/model/workflows"
)

type StepRun interface {
	StepId() string
	StepUuid() string

	Initialize(ctx context.Context, rCtx *StepRunContext) error
	PreTask() *Task
	MainTask() *Task
	PostTask() *Task
}

// ensure StepRun implementations
var (
	_ StepRun = (*ScriptStepRun)(nil)
	_ StepRun = (*DockerStepRun)(nil)
	_ StepRun = (*RepositoryStepRun)(nil)
)

type BaseStepRun struct {
	UUID             string
	ID               string
	Name             workflows.Evaluable[string]
	Condition        workflows.Conditional
	ContinueOnError  workflows.Evaluable[bool]
	TimeoutInMinutes workflows.Evaluable[int64]
	Env              workflows.Evaluable[map[string]string]
	Inputs           workflows.Evaluable[map[string]string]
}

func (s *BaseStepRun) StepId() string {
	return s.ID
}

func (s *BaseStepRun) StepUuid() string {
	return s.UUID
}

func NewStepExecutor(step workflows.Step) (StepRun, error) {
	switch s := step.(type) {
	case *workflows.RunStep:
		r := &ScriptStepRun{
			step: s,
		}
		return r, nil
	case *workflows.UsesStep:
		if image, ok := strings.CutPrefix(s.Uses, "docker://"); ok {
			e := &DockerStepRun{
				step:  s,
				Image: image,
			}
			return e, nil
		} else {
			repo, err := ParseRepository(s.Uses)
			if err != nil {
				return nil, err
			}
			e := &RepositoryStepRun{
				step: s,
				repo: repo,
			}
			return e, nil
		}
	default:
		return nil, fmt.Errorf("unknown step type: %T", step)
	}
}
