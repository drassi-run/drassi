package provision

import (
	"fmt"
	"slices"

	"drassi.run/core/pkg/sandboxer"
)

type Pipeline []Operation

func NewPipeline(ops ...Operation) Pipeline {
	return ops
}

func (p Pipeline) PreLaunch(pctx *Context) error {
	for _, op := range p {
		if err := op.PreLaunch(pctx); err != nil {
			return fmt.Errorf("operation %q pre-launch failed for runtime %q: %w", op.Name(), pctx.RuntimeName, err)
		}
	}
	return nil
}

func (p Pipeline) PostLaunch(pctx *Context, sb sandboxer.Sandbox) error {
	for _, op := range slices.Backward(p) {
		if err := op.PostLaunch(pctx, sb); err != nil {
			return fmt.Errorf("operation %q post-launch failed for runtime %q: %w", op.Name(), pctx.RuntimeName, err)
		}
	}
	return nil
}
