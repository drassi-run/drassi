package xtypes

type Pair[K, V any] struct {
	Key   K
	Value V
}

type Unwrapper[T any] interface {
	Unwrap() T
}
