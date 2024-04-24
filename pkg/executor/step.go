package executor

import (
	"context"
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/model/workflows"
)

type StepRunner interface {
	Initialize(ctx context.Context, rCtx *StepRunContext) error
	PreTask() *Task
	MainTask() *Task
	PostTask() *Task
	Step() workflows.Step
}

// ensure StepRunner implementations
var (
	_ StepRunner = (*runStepRunner)(nil)
	_ StepRunner = (*usesDockerStepRunner)(nil)
	_ StepRunner = (*usesActionStepRunner)(nil)
)

func NewStepRunner(step workflows.Step) (StepRunner, error) {
	switch s := step.(type) {
	case *workflows.RunStep:
		r := &runStepRunner{
			step: s,
		}
		return r, nil
	case *workflows.UsesStep:
		if image, ok := strings.CutPrefix(s.Uses, "docker://"); ok {
			r := &usesDockerStepRunner{
				step:  s,
				image: image,
			}
			return r, nil
		} else {
			repo, err := parseRepository(s.Uses)
			if err != nil {
				return nil, err
			}
			r := &usesActionStepRunner{
				step: s,
				repo: repo,
			}
			return r, nil
		}
	default:
		return nil, fmt.Errorf("unknown step type: %T", step)
	}
}
