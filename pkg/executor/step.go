package executor

import (
	"context"
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/model/workflows"
)

type StepExecutor interface {
	Initialize(ctx context.Context, rCtx *StepRunContext) error
	PreTask() *Task
	MainTask() *Task
	PostTask() *Task
	Step() workflows.Step
}

// ensure StepExecutor implementations
var (
	_ StepExecutor = (*runStepExecutor)(nil)
	_ StepExecutor = (*usesDockerStepExecutor)(nil)
	_ StepExecutor = (*usesActionStepExecutor)(nil)
)

func NewStepExecutor(step workflows.Step) (StepExecutor, error) {
	switch s := step.(type) {
	case *workflows.RunStep:
		r := &runStepExecutor{
			step: s,
		}
		return r, nil
	case *workflows.UsesStep:
		if image, ok := strings.CutPrefix(s.Uses, "docker://"); ok {
			e := &usesDockerStepExecutor{
				step:  s,
				image: image,
			}
			return e, nil
		} else {
			repo, err := parseRepository(s.Uses)
			if err != nil {
				return nil, err
			}
			e := &usesActionStepExecutor{
				step: s,
				repo: repo,
			}
			return e, nil
		}
	default:
		return nil, fmt.Errorf("unknown step type: %T", step)
	}
}
