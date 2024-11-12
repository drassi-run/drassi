package executor

import (
	"fmt"
	"strconv"
	"strings"

	"drassi.run/core/pkg/model/actions"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/store/repository"
	"github.com/google/uuid"
)

func ToJobRun(jobId string, job *workflows.NormalJob) *JobRun {
	stepRuns := FromSteps(job.Steps)

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

func FromSteps(steps []workflows.Step) []StepRun {
	idMap := make(map[string]int)
	stepRuns := make([]StepRun, len(steps))

	for i, step := range steps {
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
				id = normalize(s.Repo.Name)
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

	return stepRuns
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
	repo, _ := repository.Parse(s.Uses)
	return &ActionStepRun{
		BaseStepRun: *bsr,
		Repo:        repo,
	}
}

func FromAction(action *actions.Action, base BaseStepRun) (StepRun, error) {
	var sr StepRun
	switch r := action.Runs.(type) {
	case *actions.NodeRuns:
		sr = &NodeStepRun{
			BaseStepRun: base,
			Runtime:     r.Using,
			Main:        r.Main,
			Pre:         r.Pre,
			PreIf:       r.PreIf,
			Post:        r.Post,
			PostIf:      r.PostIf,
		}
	case *actions.DockerRuns:
		sr = &DockerStepRun{
			BaseStepRun:    base,
			Image:          r.Image,
			Entrypoint:     r.Entrypoint,
			Args:           r.Args,
			PreEntrypoint:  r.PreEntrypoint,
			PreIf:          r.PreIf,
			PostEntrypoint: r.PostEntrypoint,
			PostIf:         r.PostIf,
		}
		if r.Env != nil {
			if bsr := sr.Base(); bsr.Env != nil {
				bsr.Env = workflows.NewSquashMappingToken(bsr.Env, r.Env)
			} else {
				bsr.Env = r.Env
			}
		}
	case *actions.CompositeRuns:
		stepRuns := FromSteps(r.Steps)

		sr = &CompositeStepRun{
			BaseStepRun: base,
			StepRuns:    stepRuns,
		}
	default:
		return nil, fmt.Errorf("unknown action.runs %T", action.Runs)
	}

	bsr := sr.Base()

	inputTokens := make([][2]workflows.Token, 0)
	for name, input := range action.Inputs {
		if input.Default != nil {
			inputTokens = append(inputTokens, [2]workflows.Token{
				workflows.NewLiteralToken(name),
				input.Default,
			})
		} else if input.Required {
			inputTokens = append(inputTokens, [2]workflows.Token{
				workflows.NewLiteralToken(name),
				// nil value to be override
			})
		}
	}

	if len(inputTokens) > 0 {
		actionInput := workflows.NewMappingToken(inputTokens)
		if bsr.Inputs == nil {
			bsr.Inputs = actionInput
		} else {
			bsr.Inputs = workflows.NewSquashMappingToken(actionInput, bsr.Inputs)
		}
	}

	outputTokens := make([][2]workflows.Token, 0)
	for name, output := range action.Outputs {
		if output.Value != nil {
			outputTokens = append(outputTokens, [2]workflows.Token{
				workflows.NewLiteralToken(name),
				output.Value,
			})
		}
	}
	if len(outputTokens) > 0 {
		actionOutput := workflows.NewMappingToken(outputTokens)
		if bsr.Outputs == nil {
			bsr.Outputs = actionOutput
		} else {
			bsr.Outputs = workflows.NewSquashMappingToken(actionOutput, bsr.Outputs)
		}
	}

	return sr, nil
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
