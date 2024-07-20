package expression

import (
	"maps"

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
