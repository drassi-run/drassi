package xdig

import "go.uber.org/dig"

// Supply register instantiated value for dependency injection.
// It is shorthand of `scope.Provide` function where the first argument is an instance.
// See https://pkg.go.dev/go.uber.org/fx#Supply
func Supply[V any](scope *dig.Scope, value V, opts ...dig.ProvideOption) error {
	constructor := func() V {
		return value
	}
	return scope.Provide(constructor, opts...)
}

// Populate set target with value from the dependency injection container.
// It is shorthand of `scope.Invoke` function.
// See https://pkg.go.dev/go.uber.org/fx#Populate
func Populate[V any](scope *dig.Scope, target *V, opts ...dig.InvokeOption) error {
	function := func(v V) {
		*target = v
	}
	return scope.Invoke(function, opts...)
}

// Replace supplant a value in the dependency injection by a new one.
// It is shorthand of `scope.Decorate` function.
// See https://pkg.go.dev/go.uber.org/fx#Replace
func Replace[V any](scope *dig.Scope, value V, opts ...dig.DecorateOption) error {
	decorator := func() V {
		return value
	}
	return scope.Decorate(decorator, opts...)
}
