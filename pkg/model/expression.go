package model

import (
	"context"
	"strconv"
)

type Evaluable[R any] interface {
	Evaluate(context.Context) (R, error)
}

type identity[R any] struct {
	value R
}

func (i identity[R]) Evaluate(context.Context) (R, error) {
	return i.value, nil
}

// A little helper to create new identity easier
func newIdent[R any](value R) Evaluable[R] {
	return identity[R]{value: value}
}

type expression[R any] struct {
	expr      string
	converter func(string) (R, error)
}

func (e expression[R]) Evaluate(ctx context.Context) (R, error) {
	val := e.expr // TODO
	return e.converter(val)
}

// A little helper to create new expression easier
func newExpr[R any](expr string, con func(string) (R, error)) Evaluable[R] {
	return expression[R]{
		expr:      expr,
		converter: con,
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
