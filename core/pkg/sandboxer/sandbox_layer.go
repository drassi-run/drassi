package sandboxer

import (
	"context"
	"errors"
)

type layeredSandbox struct {
	Sandbox
	underlay Sandbox
}

func NewLayeredSandbox(main Sandbox, underlay Sandbox) Sandbox {
	return &layeredSandbox{
		Sandbox:  main,
		underlay: underlay,
	}
}

func (sb *layeredSandbox) Terminate(ctx context.Context) error {
	errs := make([]error, 2)
	errs[0] = sb.Sandbox.Terminate(ctx)
	errs[1] = sb.underlay.Terminate(ctx)

	return errors.Join(errs...)
}

func (sb *layeredSandbox) Underlay() Sandbox {
	return sb.underlay
}
