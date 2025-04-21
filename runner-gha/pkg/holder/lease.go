package holder

import (
	"context"

	"drassi.run/gha-runner/pkg/messages"
)

type Lease interface {
	GetMessage() *messages.PipelineAgentJobRequest
	Renew(ctx context.Context)
	Complete(ctx context.Context) error
}
