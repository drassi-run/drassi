package workflows

import (
	"context"
	"drassi.run/core/pkg/model"
	"fmt"
)

const (
	OpenExpression  = "${{"
	CloseExpression = "}}"
)

type EvaluationSupplier interface {
	Values(name string) context.Context
	Functions(name string) []string
	DefaultValue(name string) any
}

type Evaluable[R any] struct {
	Token Token `json:"token" yaml:"token" actions:"token"`
}

func (e *Evaluable[_]) IsNil() bool {
	return e == nil || e.Token == nil
}

func (e *Evaluable[R]) Evaluate(name string, supplier EvaluationSupplier) (R, error) {
	if e.IsNil() {
		v := supplier.DefaultValue(name)
		if v == nil {
			return *new(R), nil
		}

		if r, ok := v.(R); ok {
			return r, nil
		}

		return *new(R), fmt.Errorf("invalid default value for %s", name)
	}

	val, err := e.Token.Unravel(name, supplier)
	if err != nil {
		return *new(R), err
	}

	if r, ok := val.(R); ok {
		return r, nil
	}

	r := new(R)
	err = model.Decode(val, r)
	return *r, err
}

type Token interface {
	Unravel(name string, supplier EvaluationSupplier) (any, error)
}

type literalToken struct {
	value any
}

func (l *literalToken) Unravel(string, EvaluationSupplier) (any, error) {
	return l.value, nil
}

func NewLiteralToken(value any) Token {
	return &literalToken{value: value}
}

type expressionToken string

func (e expressionToken) Unravel(name string, supplier EvaluationSupplier) (any, error) {
	ctx := supplier.Values(name)
	return ctx.Value(e), nil // TODO real expression evaluation
}

func NewExpressionToken(expr string) Token {
	e := expressionToken(expr)
	return e
}

type sequenceToken []Token

func (s sequenceToken) Unravel(name string, supplier EvaluationSupplier) (any, error) {
	r := make([]any, len(s))

	for i, token := range s {
		if e, err := token.Unravel(name, supplier); err != nil {
			return nil, err
		} else {
			r[i] = e
		}
	}
	return r, nil
}

func NewSequenceToken(seq []Token) Token {
	e := sequenceToken(seq)
	return e
}

type mappingToken [][2]Token

func (m mappingToken) Unravel(name string, supplier EvaluationSupplier) (any, error) {
	r := make(map[string]any, len(m))

	for _, pair := range m {
		kAny, err := pair[0].Unravel(name, supplier)
		if err != nil {
			return nil, err
		}
		k, ok := kAny.(string)
		if !ok {
			return nil, fmt.Errorf("invalid key type: %T", kAny)
		}

		v, err := pair[1].Unravel(name+"."+k, supplier)
		if err != nil {
			return nil, err
		}

		r[k] = v
	}
	return r, nil
}

func NewMappingToken(pairs [][2]Token) Token {
	e := mappingToken(pairs)
	return e
}
