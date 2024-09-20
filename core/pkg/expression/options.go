package expression

import (
	"strings"

	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
)

type Option func(o *env) error

func WithAlias(name string, ref string) Option {
	return func(o *env) error {
		o.alias[name] = ref
		return nil
	}
}

func WithVariable(name string, value any) Option {
	return func(o *env) error {
		variable := types.NativeToVal(value)
		if ref.IsError(variable) {
			return variable.Value().(error)
		}

		o.variables[name] = variable
		return nil
	}
}

func WithFunction(name string, fn Function) Option {
	// function name is case-insensitive
	name = strings.ToLower(name)

	return func(o *env) error {
		o.functions[name] = fn
		return nil
	}
}

type Function interface {
	NumArgs() (min int, max int)
	Bind(args ...ref.LazyVal) ref.LazyVal
}

func WithLibrary(lib Library) Option {
	return func(o *env) error {
		for _, opt := range lib.EnvOptions() {
			if err := opt(o); err != nil {
				return err
			}
		}
		return nil
	}
}

type Library interface {
	EnvOptions() []Option
}

func WithMaxError(num int) Option {
	return func(o *env) error {
		o.maxError = num
		return nil
	}
}

func WithMaxDepth(depth int) Option {
	return func(o *env) error {
		o.maxDepth = depth
		return nil
	}
}

func WithMaxLength(length int) Option {
	return func(o *env) error {
		o.maxLength = length
		return nil
	}
}

func WithCache(enabled bool) Option {
	return func(o *env) error {
		o.cacheEnabled = enabled
		return nil
	}
}
