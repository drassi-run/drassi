/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

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
			switch def := spec.Def.(type) {
			case *ScriptStepDef:
				id = "run"
			case *DockerStepDef:
				id = normalize(def.Image)
			case *ActionStepDef:
				id = normalize(def.Repo.Name)
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
		spec.Def = &ScriptStepDef{
			Run:        s.Run,
			Shell:      s.Shell,
			WorkingDir: s.WorkingDir,
		}
	case *workflows.UsesStep:
		spec.Inputs = s.With
		if strings.HasPrefix(s.Uses, "docker://") {
			spec.Def = &DockerStepDef{
				Image: s.Uses,
			}
		} else {
			repo, _ := repository.Parse(s.Uses)
			spec.Def = &ActionStepDef{
				Repo: repo,
			}
		}
	}
	return spec
}

func ToStepDef(action *actions.Action) (StepDef, error) {
	var def StepDef
	switch r := action.Runs.(type) {
	case *actions.NodeRuns:
		def = &NodeStepDef{
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
		def = &DockerStepDef{
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

		def = &CompositeStepDef{
			Inputs:  inputToken(action.Inputs),
			Outputs: outputToken(action.Outputs),

			Steps: stepRuns,
		}
	default:
		return nil, fmt.Errorf("unknown action.runs %T", action.Runs)
	}

	return def, nil
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
