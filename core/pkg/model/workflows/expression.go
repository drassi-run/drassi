package workflows

import (
	"fmt"

	"github.com/dungdm93/drassi/core/pkg/executor/secret"
	"github.com/dungdm93/drassi/core/pkg/expression/evaluator"
	"github.com/dungdm93/drassi/core/pkg/expression/parser"
	"github.com/dungdm93/drassi/core/pkg/expression/tools"
	"github.com/dungdm93/drassi/core/pkg/model"
	"github.com/dungdm93/drassi/core/pkg/model/contexts"
)

const (
	OpenExpression  = "${{"
	CloseExpression = "}}"
)

type EvaluatorProvider interface {
	ContextData(name string) contexts.Context
	Functions(name string) []string
	DefaultValue(name string) any
}

type Evaluable[R any] struct {
	Token Token `json:"token" yaml:"token" mapstructure:"token"`
}

func (e *Evaluable[R]) Evaluate(name string, provider EvaluatorProvider) (R, error) {
	if e.Token == nil {
		v := provider.DefaultValue(name)
		if v == nil {
			return *new(R), nil
		}

		if r, ok := v.(R); ok {
			return r, nil
		}

		return *new(R), fmt.Errorf("invalid default value for %s", name)
	}

	val, err := e.Token.Appraise(name, provider)
	if err != nil {
		return *new(R), err
	}

	if r, ok := val.(R); ok {
		return r, nil
	}

	r := new(R)
	err = model.Decode(val, r)
	if err != nil {
		return *r, err
	}

	return *r, nil
}

type Token interface {
	Appraise(name string, provider EvaluatorProvider) (any, error)
}

type literalToken struct {
	value any
}

func (l *literalToken) Appraise(string, EvaluatorProvider) (any, error) {
	return l.value, nil
}

func NewLiteralToken(value any) Token {
	return &literalToken{value: value}
}

type expressionToken string

func (e *expressionToken) Appraise(name string, provider EvaluatorProvider) (any, error) {
	ctx := provider.ContextData(name)
	functionNames := provider.Functions(name)
	ast := parser.ParseWithDefaults(string(*e),  tools.GetFunctionNode(functionNames))
	ret, err := evaluator.Evaluate(ast, new(secret.Masker), ctx, nil)
	if err != nil {
		return nil, err
	}
	return ret.Value(), nil
}

func NewExpressionToken(expr string) Token {
	e := expressionToken(expr)
	return &e
}

type sequenceToken []Token

func (s *sequenceToken) Appraise(name string, provider EvaluatorProvider) (any, error) {
	seq := []Token(*s)
	r := make([]any, len(seq))

	for i, token := range seq {
		if e, err := token.Appraise(name, provider); err != nil {
			return nil, err
		} else {
			r[i] = e
		}
	}
	return r, nil
}

func NewSequenceToken(seq []Token) Token {
	e := sequenceToken(seq)
	return &e
}

type KVPair[K, V any] struct {
	Key   K `json:"key,omitempty" yaml:"key,omitempty" mapstructure:"key,omitempty"`
	Value V `json:"value,omitempty" yaml:"value,omitempty" mapstructure:"value,omitempty"`
}

type mappingToken []KVPair[Token, Token]

func (m *mappingToken) Appraise(name string, provider EvaluatorProvider) (any, error) {
	pairs := []KVPair[Token, Token](*m)
	r := make(map[string]any, len(pairs))

	for _, pair := range pairs {
		kAny, err := pair.Key.Appraise(name, provider)
		if err != nil {
			return nil, err
		}
		k, ok := kAny.(string)
		if !ok {
			return nil, fmt.Errorf("invalid key type: %T", kAny)
		}

		v, err := pair.Value.Appraise(name+"."+k, provider)
		if err != nil {
			return nil, err
		}

		r[k] = v
	}
	return r, nil
}

func NewMappingToken(pairs []KVPair[Token, Token]) Token {
	e := mappingToken(pairs)
	return &e
}
