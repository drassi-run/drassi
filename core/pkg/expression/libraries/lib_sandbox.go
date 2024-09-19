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
	return &sandboxLib{contextual: c, sandbox: sb}
}

type sandboxLib struct {
	contextual Contextual
	sandbox    sandboxer.Sandbox
}

func (s *sandboxLib) EnvOptions() []expr.Option {
	// TODO add hashFiles function
	return nil
}
