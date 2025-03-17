package xtypes

import "context"

type Pair[K, V any] struct {
	Key   K
	Value V
}

type Unwrapper[T any] interface {
	Unwrap() T
}

type ContextProvider interface {
	Context() context.Context
}

func NewStaticContext(ctx context.Context) ContextProvider {
	return &staticContextProvider{ctx: ctx}
}

type staticContextProvider struct {
	ctx context.Context
}

func (p *staticContextProvider) Context() context.Context {
	return p.ctx
}
