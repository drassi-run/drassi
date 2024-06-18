package converter

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drassi/core/pkg/executor"
	"github.com/dungdm93/drassi/core/pkg/model"
	"github.com/dungdm93/drassi/core/pkg/model/workflows"
	"github.com/dungdm93/drassi/gha-runner/pkg/message"
)

func ToToken(token *message.TemplateToken) workflows.Token {
	if token == nil {
		return nil
	}

	switch token.Type {
	case message.TokenTypeString:
		return workflows.NewLiteralToken(token.String)
	case message.TokenTypeSequence:
		seq := make([]workflows.Token, len(token.Seq))
		for i, s := range token.Seq {
			seq[i] = ToToken(s)
		}
		return workflows.NewSequenceToken(seq)
	case message.TokenTypeMapping:
		pairs := make([]workflows.KVPair[workflows.Token, workflows.Token], len(token.Map))
		for i, m := range token.Map {
			k := ToToken(m.Key)
			v := ToToken(m.Value)
			pairs[i] = workflows.KVPair[workflows.Token, workflows.Token]{
				Key:   k,
				Value: v,
			}
		}
		return workflows.NewMappingToken(pairs)
	case message.TokenTypeBasicExpression:
		return workflows.NewExpressionToken(token.Expr)
	case message.TokenTypeInsertExpression:
		// TODO: not supported
		return nil
	case message.TokenTypeNumber:
		return workflows.NewLiteralToken(token.Number)
	case message.TokenTypeBoolean:
		return workflows.NewLiteralToken(token.Boolean)
	case message.TokenTypeNull:
		return nil
	}
	return nil
}

func ToEvaluable[R any](token *message.TemplateToken) workflows.Evaluable[R] {
	return workflows.Evaluable[R]{
		Token: ToToken(token),
	}
}

func squashTokens(tokens []message.TemplateToken) workflows.Token {
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
		return &multiMapToken{tokens: ts}
	}
}

func ToStepRun(step *message.JobStep) (executor.StepRun, error) {
	sr := executor.BaseStepRun{
		Uid:              step.Id,
		Id:               step.ContextName,
		Name:             ToEvaluable[string](step.DisplayNameToken),
		ContinueOnError:  ToEvaluable[bool](step.ContinueOnError),
		TimeoutInMinutes: ToEvaluable[int64](step.TimeoutInMinutes),
		Env:              ToEvaluable[map[string]string](step.Env),
		Inputs:           ToEvaluable[map[string]string](step.Inputs),
	}
	if step.Condition != "" {
		sr.Condition = workflows.NewConditional(step.Condition)
	}

	ref := &step.Reference
	switch ref.Type {
	case message.SourceTypeScript:
		ssr := &executor.ScriptStepRun{
			BaseStepRun: sr,
		}
		return ssr, nil
	case message.SourceTypeContainerRegistry:
		if ref.Image == "" {
			return nil, fmt.Errorf("step %s image is required", step.ContextName)
		}
		dsr := &executor.DockerStepRun{
			BaseStepRun: sr,
			Image:       ref.Image,
		}
		return dsr, nil
	case message.SourceTypeRepository:
		if !strings.EqualFold(ref.RepositoryType, "github") {
			return nil, fmt.Errorf("unsupported step %s with repo type %s", step.ContextName, ref.RepositoryType)
		}
		repo := model.Repository{
			Protocol: "https",
			Endpoint: "github.com",
			Repo:     ref.Name,
			Path:     ref.Path,
			Ref:      ref.Ref,
		}
		rsr := &executor.RepositoryStepRun{
			BaseStepRun: sr,
			Repo:        repo,
		}
		return rsr, nil
	default:
		return nil, fmt.Errorf("unsupported step %s reference type %s", step.ContextName, ref.Type)
	}
}

func ToJobRun(job *message.PipelineAgentJobRequest) (*executor.JobRun, error) {
	steps := make([]executor.StepRun, len(job.Steps))
	for i, s := range job.Steps {
		step, err := ToStepRun(&s)
		if err != nil {
			return nil, err
		}
		steps[i] = step
	}

	jr := &executor.JobRun{
		Uid: job.JobId,
		Id:  job.JobName,
		Name: workflows.Evaluable[string]{
			Token: workflows.NewLiteralToken(job.JobDisplayName),
		},

		Container: ToEvaluable[*workflows.Container](job.JobContainer),
		Services:  ToEvaluable[map[string]*workflows.Container](job.JobServiceContainers),

		Defaults: workflows.Evaluable[workflows.Defaults]{
			Token: squashTokens(job.Defaults),
		},
		Env: workflows.Evaluable[map[string]string]{
			Token: squashTokens(job.Env),
		},
		Steps:   steps,
		Outputs: ToEvaluable[map[string]string](job.JobOutputs),
	}
	return jr, nil
}
