package expression

import (
	"fmt"
	"maps"

	"drassi.run/core/pkg/expression/ast"
	"drassi.run/core/pkg/expression/types/ref"
)

type Env struct {
	variables map[string]ref.Val
	functions map[string]Function
}

func NewEnv(base *Env, opts ...EnvOption) (*Env, error) {
	o := &option{
		variables: make(map[string]ref.Val),
		functions: make(map[string]Function),
	}
	if base != nil {
		o.variables = maps.Clone(base.variables)
		o.functions = maps.Clone(base.functions)
	}

	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}

	e := &Env{
		variables: o.variables,
		functions: o.functions,
	}
	return e, nil
}

func (e *Env) Bind(node ast.Node) (prog ref.LazyVal, err error) {
	defer e.recover(&err)

	b := binder{Env: e}
	return b.Bind(node)
}

func (e *Env) Execute(prog ref.LazyVal) (result any, err error) {
	defer e.recover(&err)

	val := prog()
	if err, ok := val.(error); ok {
		return nil, err
	}
	return val.Value(), nil
}

func (e *Env) recover(err *error) {
	if r := recover(); r != nil {
		ex, ok := r.(error)
		if !ok {
			ex = fmt.Errorf("panic: %v", r)
		}
		*err = ex
	}
}
