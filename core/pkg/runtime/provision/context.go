package provision

import (
	"context"
	"fmt"
	"sync"

	"drassi.run/core/config"
)

type StateKey[T any] string

type Context struct {
	context.Context

	RuntimeName string
	Config      *config.Runtime
	TargetDir   string

	state sync.Map
}

func NewContext(ctx context.Context, name string, cfg *config.Runtime, targetDir string) *Context {
	return &Context{
		Context:     ctx,
		RuntimeName: name,
		Config:      cfg,
		TargetDir:   targetDir,
	}
}

func (c *Context) Set[T any](key StateKey[T], val T) {
	c.state.Store(string(key), val)
}

func (c *Context) Get[T any](key StateKey[T]) (T, bool) {
	if v, ok := c.state.Load(string(key)); ok {
		if val, ok := v.(T); ok {
			return val, true
		}
	}
	var zero T
	return zero, false
}

func (c *Context) MustGet[T any](key StateKey[T]) T {
	if val, ok := c.Get(key); ok {
		return val
	}
	panic(fmt.Sprintf("state key %q not found or type mismatch", string(key)))
}
