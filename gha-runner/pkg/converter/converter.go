package converter

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drassi/core/pkg/executor"
	"github.com/dungdm93/drassi/core/pkg/model/workflows"
	"github.com/dungdm93/drassi/gha-runner/pkg/message"
)

func ToToken(token *message.TemplateToken) workflows.Token {
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

func ToEnv(tokens []message.TemplateToken) workflows.Evaluable[map[string]string] {
	switch len(tokens) {
	case 0:
		return workflows.Evaluable[map[string]string]{}
	case 1:
		return workflows.Evaluable[map[string]string]{
			Token: ToToken(&tokens[0]),
		}
	default:
		ts := make([]workflows.Token, len(tokens))
		for i, token := range tokens {
			ts[i] = ToToken(&token)
		}
		return workflows.Evaluable[map[string]string]{
			Token: &multiMapToken{tokens: ts},
		}
	}
}

func ToStepRun(step *message.JobStep) (executor.StepRun, error) {
	sr := executor.BaseStepRun{
		UUID:             step.Id,
		ID:               step.ContextName,
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
		repo := executor.Repository{
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
