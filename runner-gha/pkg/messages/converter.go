/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package messages

import (
	"fmt"
	"strings"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/store/repository"
)

func ToToken(token *TemplateToken) workflows.Token {
	if token == nil {
		return nil
	}

	switch token.Type {
	case TokenTypeString:
		return workflows.NewLiteralToken(token.String)
	case TokenTypeSequence:
		seq := make([]workflows.Token, len(token.Seq))
		for i, s := range token.Seq {
			seq[i] = ToToken(s)
		}
		return workflows.NewSequenceToken(seq)
	case TokenTypeMapping:
		pairs := make([][2]workflows.Token, len(token.Map))
		for i, m := range token.Map {
			k := ToToken(m.Key)
			v := ToToken(m.Value)

			pairs[i] = [2]workflows.Token{k, v}
		}
		return workflows.NewMappingToken(pairs)
	case TokenTypeBasicExpression:
		return workflows.NewExpressionToken(token.Expr)
	case TokenTypeInsertExpression:
		// TODO: not supported
		return nil
	case TokenTypeNumber:
		return workflows.NewLiteralToken(token.Number)
	case TokenTypeBoolean:
		return workflows.NewLiteralToken(token.Boolean)
	case TokenTypeNull:
		return nil
	}
	return nil
}

func squashTokens(tokens []TemplateToken) workflows.Token {
	switch len(tokens) {
	case 0:
		return nil
	case 1:
		return ToToken(&tokens[0])
	default:
		ts := make([]workflows.Token, len(tokens))
		for i, token := range tokens {
			ts[i] = ToToken(&token)
		}
		return workflows.NewSquashMappingToken(ts...)
	}
}

func ToStepRun(step *JobStep) (executor.StepSpec, error) {
	sr := executor.BaseStepRun{
		Uid:              step.Id,
		Id:               step.ContextName,
		Name:             ToToken(step.DisplayNameToken),
		Condition:        workflows.Conditional(step.Condition),
		ContinueOnError:  ToToken(step.ContinueOnError),
		TimeoutInMinutes: ToToken(step.TimeoutInMinutes),
		Env:              ToToken(step.Env),
		Inputs:           ToToken(step.Inputs),
	}
	// for Script step, extract run, shell and workingDir from inputs
	if step.Reference.Type != SourceTypeScript {
		sr.Inputs = ToToken(step.Inputs)
	}

	ref := &step.Reference
	switch ref.Type {
	case SourceTypeScript:
		ssr := &executor.ScriptActionSpec{
			BaseStepRun: sr,
		}
		if err := extractScriptStepInputs(ssr, step.Inputs); err != nil {
			return nil, err
		}
		return ssr, nil
	case SourceTypeContainerRegistry:
		if ref.Image == "" {
			return nil, fmt.Errorf("step %s image is required", step.ContextName)
		}
		dsr := &executor.DockerActionSpec{
			BaseStepRun: sr,
			Image:       ref.Image,
		}
		return dsr, nil
	case SourceTypeRepository:
		if !strings.EqualFold(ref.RepositoryType, "github") {
			return nil, fmt.Errorf("unsupported step %s with repo type %s", step.ContextName, ref.RepositoryType)
		}
		repo := &repository.Repository{
			Scheme:   "https",
			Endpoint: "github.com",
			Name:     ref.Name,
			Path:     ref.Path,
			Ref:      ref.Ref,
		}
		asr := &executor.ReferenceActionSpec{
			BaseStepRun: sr,
			Repo:        repo,
		}
		return asr, nil
	default:
		return nil, fmt.Errorf("unsupported step %s reference type %s", step.ContextName, ref.Type)
	}
}

func ToJobSpec(job *PipelineAgentJobRequest) (*executor.JobSpec, error) {
	steps := make([]executor.StepSpec, len(job.Steps))
	for i, s := range job.Steps {
		step, err := ToStepRun(&s)
		if err != nil {
			return nil, err
		}
		steps[i] = step
	}

	spec := &executor.JobSpec{
		Uid:  job.JobId,
		Id:   job.JobName,
		Name: workflows.NewLiteralToken(job.JobDisplayName),

		Container: ToToken(job.JobContainer),
		Services:  ToToken(job.JobServiceContainers),

		Defaults: squashTokens(job.Defaults),
		Env:      squashTokens(job.Env),
		Steps:    steps,
		Outputs:  ToToken(job.JobOutputs),
	}
	return spec, nil
}

func extractScriptStepInputs(ssr *executor.ScriptActionSpec, inputs *TemplateToken) error {
	if inputs.Type != TokenTypeMapping {
		return fmt.Errorf("exptect step inputs is a map, got %d", inputs.Type)
	}

	for _, pair := range inputs.Map {
		k := pair.Key
		v := pair.Value
		if k.Type != TokenTypeString {
			return fmt.Errorf("exptect step inputs key is a string, got %d", k.Type)
		}
		switch k.String {
		case "script":
			ssr.Run = ToToken(v)
		case "workingDirectory":
			ssr.WorkingDir = ToToken(v)
		case "shell":
			if v.Type != TokenTypeString {
				return fmt.Errorf("exptect step shell is a string, got %d", k.Type)
			}
			ssr.Shell = v.String
		default:
			return fmt.Errorf("unexppected step inputs key %s", k.String)
		}
	}

	return nil
}
