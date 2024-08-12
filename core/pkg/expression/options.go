package expression

import (
	"strings"

	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
)

type option struct {
	variables map[string]ref.Val
	functions map[string]Function
}

type EnvOption func(o *option) error

func WithVariable(name string, value any) EnvOption {
	return func(o *option) error {
		variable := types.NativeToVal(value)
		if variable.Type() == ref.TypeInvalid {
			return variable.Value().(error)
		}

		o.variables[name] = variable
		return nil
	}
}

func WithFunction(name string, fn Function) EnvOption {
	// function name is case-insensitive
	name = strings.ToLower(name)

	return func(o *option) error {
		o.functions[name] = fn
		return nil
	}
}

type Function interface {
	NumArgs() (min int, max int)
	Bind(args ...ref.LazyVal) ref.LazyVal
}

func WithLibrary(lib Library) EnvOption {
	return func(o *option) error {
		for _, opt := range lib.EnvOptions() {
			if err := opt(o); err != nil {
				return err
			}
		}
		return nil
	}
}

type Library interface {
	EnvOptions() []EnvOption
}
