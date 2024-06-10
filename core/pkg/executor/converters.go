package executor

import (
	"context"

	"github.com/dungdm93/drassi/core/pkg/container"
	"github.com/dungdm93/drassi/core/pkg/executor/problem"
	"github.com/dungdm93/drassi/core/pkg/executor/reporter"
	"github.com/dungdm93/drassi/core/pkg/model/workflows"
)

func (e *JobExecutor) toContainerConfig(ctx context.Context, container *workflows.Container) (*container.ContainerConfig, error) {
	return nil, nil
}

func (e *JobExecutor) toIssuer(pbl *problem.Problem) (*reporter.Issue, error) {

}
