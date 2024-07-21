package expression

import (
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

func (e *Env) newBinder() *binder {
	return &binder{Env: e}
}

func (e *Env) Evaluate(node ast.Node) (any, error) {
	b := e.newBinder()

	if prog, err := b.Bind(node); err != nil {
		return nil, err
	} else {
		val := prog()
		if err, ok := val.(error); ok {
			return nil, err
		}
		return val.Value(), nil
	}
}
