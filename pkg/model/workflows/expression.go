package workflows

import (
	"context"
	"strconv"
	"strings"
)

type Evaluable[R any] interface {
	Evaluate(context.Context) (R, error)
}

type identity[R any] struct {
	value R
}

func (i *identity[R]) Evaluate(context.Context) (R, error) {
	return i.value, nil
}

// A little helper to create new identity easier
func NewIdent[R any](value R) Evaluable[R] {
	return &identity[R]{value: value}
}

type converter[R any] func(string) (R, error)
type expression[R any] struct {
	expr string
	conv converter[R]
}

func (e *expression[R]) Evaluate(ctx context.Context) (R, error) {
	val := e.expr // TODO
	return e.conv(val)
}

func (e *expression[bool]) Meet(ctx context.Context) (bool, error) {
	return e.Evaluate(ctx)
}

// A little helper to create new expression easier
func NewExpr[R any](expr string, conv converter[R]) Evaluable[R] {
	return &expression[R]{
		expr: expr,
		conv: conv,
	}
}

var toBool = strconv.ParseBool

func toInteger(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func toFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func toString(s string) (string, error) {
	return s, nil
}

func toMatrix(s string) (Matrix, error) {
	return Matrix{}, nil // TODO
}

const (
	OpenExpression  = "${{"
	CloseExpression = "}}"
)

func NewEvaluable[R any](s string, con converter[R]) (Evaluable[R], error) {
	if strings.Contains(s, OpenExpression) {
		return NewExpr(s, con), nil
	}

	if v, err := con(s); err != nil {
		return nil, err
	} else {
		return NewIdent(v), nil
	}
}

// Conditional is Evaluable[bool] type used by `if`, `pre-if` and `post-if`.
// The `${{ }}` expression syntax is optional and can be omitted. GitHub Actions always evaluates it as an expression.
type Conditional interface {
	Meet(context.Context) (bool, error)
}

func NewConditional(s string) Conditional {
	return &expression[bool]{
		expr: s,
		conv: toBool,
	}
}
