package libraries

import (
	"context"
	expr "drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/sandboxer"
)

type ContextProvider interface {
	Context() context.Context
}

func SandboxLib(c ContextProvider, sb sandboxer.Sandbox) expr.Library {
	return &sandboxLib{contextual: c, sandbox: sb}
}

type sandboxLib struct {
	contextual ContextProvider
	sandbox    sandboxer.Sandbox
}

func (s *sandboxLib) EnvOptions() []expr.Option {
	// TODO add hashFiles function
	return nil
}
