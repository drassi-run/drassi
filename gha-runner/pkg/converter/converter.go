package converter

import (
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
