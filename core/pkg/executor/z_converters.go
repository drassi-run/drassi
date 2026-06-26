/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package executor

import (
	"fmt"
	"maps"
	"strconv"
	"strings"

	"drassi.run/core/pkg/model/actions"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/store/repository"
	"github.com/google/uuid"
)

func ToJobSpec(jobId string, job *workflows.NormalJob) *JobSpec {
	steps := fromSteps(job.Steps)

	uid, _ := uuid.NewRandom()
	return &JobSpec{
		Id:        jobId,
		Uid:       uid.String(),
		Name:      job.Name,
		Container: job.Container,
		Services:  job.Services,
		Env:       job.Env,
		Steps:     steps,
		Outputs:   job.Outputs,
		Defaults:  job.Defaults,
	}
}

func fromSteps(steps []workflows.Step) []*StepSpec {
	idMap := make(map[string]int)
	specs := make([]*StepSpec, len(steps))

	for i, step := range steps {
		spec := ToStepSpec(step)

		// generate StepId if empty
		if spec.Id == "" {
			var id string
			switch action := spec.Action.(type) {
			case *ScriptActionSpec:
				id = "run"
			case *DockerActionSpec:
				id = normalize(action.Image)
			case *ReferenceActionSpec:
				id = normalize(action.Repo.Name)
			}

			count := idMap[id] + 1
			idMap[id] = count
			if count > 1 {
				id += "_" + strconv.Itoa(count)
			}

			spec.Id = "__" + id
		}

		specs[i] = spec
	}

	return specs
}

func ToStepSpec(step workflows.Step) *StepSpec {
	b := step.Base()
	uid, _ := uuid.NewRandom()
	spec := &StepSpec{
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
		spec.Action = &ScriptActionSpec{
			Run:        s.Run,
			Shell:      s.Shell,
			WorkingDir: s.WorkingDir,
		}
	case *workflows.UsesStep:
		spec.Inputs = s.With
		if strings.HasPrefix(s.Uses, "docker://") {
			spec.Action = &DockerActionSpec{
				Image: s.Uses,
			}
		} else {
			repo, _ := repository.Parse(s.Uses)
			spec.Action = &ReferenceActionSpec{
				Repo: repo,
			}
		}
	}
	return spec
}

func ToActionSpec(action *actions.Action, repo *repository.Repository) (ActionSpec, error) {
	var spec ActionSpec
	switch r := action.Runs.(type) {
	case *actions.NodeRuns:
		spec = &NodeActionSpec{
			Repo:    repo,
			Inputs:  inputToken(action.Inputs),
			Outputs: outputToken(action.Outputs),

			Runtime: r.Using,
			Main:    r.Main,
			Pre:     r.Pre,
			PreIf:   r.PreIf,
			Post:    r.Post,
			PostIf:  r.PostIf,
		}
	case *actions.DockerRuns:
		spec = &DockerActionSpec{
			Repo:    repo,
			Inputs:  inputToken(action.Inputs),
			Outputs: outputToken(action.Outputs),
			Env:     r.Env,

			Image:          r.Image,
			Entrypoint:     r.Entrypoint,
			Args:           r.Args,
			PreEntrypoint:  r.PreEntrypoint,
			PreIf:          r.PreIf,
			PostEntrypoint: r.PostEntrypoint,
			PostIf:         r.PostIf,
		}
	case *actions.CompositeRuns:
		stepRuns := fromSteps(r.Steps)

		spec = &CompositeActionSpec{
			Repo:    repo,
			Inputs:  inputToken(action.Inputs),
			Outputs: outputToken(action.Outputs),

			Steps: stepRuns,
		}
	default:
		return nil, fmt.Errorf("unknown action.runs %T", action.Runs)
	}

	return spec, nil
}

func inputToken(m map[string]workflows.Input) workflows.Token {
	tokens := make([][2]workflows.Token, 0)
	for name, input := range m {
		if input.Default != nil {
			tokens = append(tokens, [2]workflows.Token{
				workflows.NewLiteralToken(name),
				input.Default,
			})
		} else if input.Required {
			tokens = append(tokens, [2]workflows.Token{
				workflows.NewLiteralToken(name),
				// nil value to be overridden
			})
		}
	}
	if len(tokens) > 0 {
		return workflows.NewMappingToken(tokens)
	}
	return nil
}

func outputToken(m map[string]workflows.Output) workflows.Token {
	tokens := make([][2]workflows.Token, 0)
	for name, output := range m {
		if output.Value != nil {
			tokens = append(tokens, [2]workflows.Token{
				workflows.NewLiteralToken(name),
				output.Value,
			})
		}
	}
	if len(tokens) > 0 {
		return workflows.NewMappingToken(tokens)
	}
	return nil
}

func mergeMapExpr(a, b workflows.Evaluable[map[string]string]) workflows.Evaluable[map[string]string] {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	return workflows.NewSquashMappingToken(a, b)
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

func composeEnv(exec StepExecutor) map[string]string {
	env := exec.SystemEnv()
	maps.Copy(env, exec.Env())
	return env
}

func weight(r records.Result) int {
	switch r {
	case records.ResultSkipped:
		return 0
	case records.ResultSuccess:
		return 1
	case records.ResultCancelled:
		return 2
	case records.ResultFailure:
		return 3
	default:
		return 0
	}
}
