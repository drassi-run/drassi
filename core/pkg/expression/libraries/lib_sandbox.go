package libraries

import (
	"context"
	expr "drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/sandboxer"
)

type Contextual interface {
	Context() context.Context
}

func SandboxLib(c Contextual, sb sandboxer.Sandbox) expr.Library {
	lib := &sandboxLib{contextual: c, sandbox: sb}

	return lib
}

type sandboxLib struct {
	contextual Contextual
	sandbox    sandboxer.Sandbox
}

func (s *sandboxLib) EnvOptions() []expr.EnvOption {
	// TODO add hashFiles function
	return nil
}
