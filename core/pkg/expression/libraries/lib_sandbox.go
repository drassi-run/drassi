package libraries

import (
	expr "drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/util/types"
)

func SandboxLib(c xtypes.ContextProvider, sb sandboxer.Sandbox) expr.Library {
	return &sandboxLib{contextual: c, sandbox: sb}
}

type sandboxLib struct {
	contextual xtypes.ContextProvider
	sandbox    sandboxer.Sandbox
}

func (s *sandboxLib) EnvOptions() []expr.Option {
	// TODO add hashFiles function
	return nil
}
