package provision

import (
	"fmt"

	"drassi.run/core/pkg/sandboxer"
)

type Operation interface {
	Name() string
	PreLaunch(pctx *Context) error
	PostLaunch(pctx *Context, sb sandboxer.Sandbox) error
}

type Pipeline struct {
	ops []Operation
}

func NewPipeline(ops ...Operation) *Pipeline {
	return &Pipeline{ops: ops}
}

func (p *Pipeline) PreLaunch(pctx *Context) error {
	for _, op := range p.ops {
		if err := op.PreLaunch(pctx); err != nil {
			return fmt.Errorf("operation %q pre-launch failed for runtime %q: %w", op.Name(), pctx.RuntimeName, err)
		}
	}
	return nil
}

func (p *Pipeline) PostLaunch(pctx *Context, sb sandboxer.Sandbox) error {
	for _, op := range p.ops {
		if err := op.PostLaunch(pctx, sb); err != nil {
			return fmt.Errorf("operation %q post-launch failed for runtime %q: %w", op.Name(), pctx.RuntimeName, err)
		}
	}
	return nil
}
