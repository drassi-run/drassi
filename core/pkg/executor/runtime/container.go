package runtime

import (
	"context"
)

// Container runtime is used to run docker action
type Container interface {
	// TranslatePath map from containerPath to sandboxPath,
	TranslatePath(containerPath string) (sandboxPath string, ok bool)

	Pull(ctx context.Context, image string) error
	Build(ctx context.Context) error
	Run(ctx context.Context, image string, entrypoint, cmd []string, env map[string]string) error
}
