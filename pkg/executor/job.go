package executor

import (
	"context"
	"github.com/dungdm93/drasi/pkg/container"
	"github.com/dungdm93/drasi/pkg/model/workflows"
	"github.com/dungdm93/drasi/pkg/sandboxer"
)

type Stage string

const (
	StagePre  Stage = "pre"
	StageMain Stage = "main"
	StagePost Stage = "post"
)

type Task struct {
	StepID    string
	Stage     Stage
	Condition workflows.Conditional // default true
	Run       func(context.Context, *StepRunContext) error
}

type JobRunner struct {
	job       *workflows.NormalJob
	jobId     string
	sandboxer sandboxer.Sandboxer

	rCtx *JobRunContext
}

func NewJobRunner(job *workflows.NormalJob, jobId string, sandboxer sandboxer.Sandboxer) *JobRunner {
	return &JobRunner{
		job:       job,
		jobId:     jobId,
		sandboxer: sandboxer,
	}
}

func (e *JobRunner) Initialize(ctx context.Context) (err error) {
	e.rCtx = &JobRunContext{
		job: e.job,
	}

	var jobContainer *container.ContainerConfig
	if e.job.Container != nil {
		jobContainer, err = toContainerConfig(ctx, e.rCtx, e.job.Container)
		if err != nil {
			return err
		}
	}
	var serviceContainers = make(map[string]*container.ContainerConfig)
	for name, con := range e.job.Services {
		serviceContainers[name], err = toContainerConfig(ctx, e.rCtx, con)
		if err != nil {
			return err
		}
	}

	req := sandboxer.LaunchSandboxRequest{
		JobId:             e.jobId,
		JobEnv:            e.rCtx.env,
		JobContainer:      jobContainer,
		ServiceContainers: serviceContainers,
	}
	if res, err := e.sandboxer.LaunchSandbox(ctx, req); err != nil {
		return err
	} else {
		e.rCtx.sandbox = res.Sandbox
	}
	return nil
}

func (e *JobRunner) Run(ctx context.Context) error {
	return e.rCtx.RunJob(ctx)
}

func (e *JobRunner) Finalize(ctx context.Context) error {
	req := sandboxer.TerminateSandboxRequest{
		Sandbox: e.rCtx.sandbox,
	}
	_, err := e.sandboxer.TerminateSandbox(ctx, req)
	return err
}

func toContainerConfig(ctx context.Context, rCtx *JobRunContext, container *workflows.Container) (*container.ContainerConfig, error) {
	return nil, nil
}
