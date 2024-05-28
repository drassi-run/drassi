package executor

import (
	"context"
	"github.com/dungdm93/drassi/core/pkg/container"
	"github.com/dungdm93/drassi/core/pkg/model/workflows"
	"github.com/dungdm93/drassi/core/pkg/sandboxer"
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

type JobExecutor struct {
	job   *workflows.NormalJob
	jobId string

	rCtx *JobRunContext
}

func NewJobExecutor(job *workflows.NormalJob, jobId string) *JobExecutor {
	return &JobExecutor{
		job:   job,
		jobId: jobId,
	}
}

func (e *JobExecutor) Initialize(ctx context.Context, runtime sandboxer.SandboxRuntime) error {
	e.rCtx = &JobRunContext{
		job: e.job,
	}

	var jobContainer *container.ContainerConfig
	if con, err := e.job.Container.Evaluate("job.container", e.rCtx); err != nil {
		return err
	} else {
		jobContainer, err = toContainerConfig(ctx, e.rCtx, con)
		if err != nil {
			return err
		}
	}
	var serviceContainers = make(map[string]*container.ContainerConfig)
	if sers, err := e.job.Services.Evaluate("job.services", e.rCtx); err != nil {
		return err
	} else {
		for name, ser := range sers {
			con, err := toContainerConfig(ctx, e.rCtx, ser)
			if err != nil {
				return err
			}
			serviceContainers[name] = con
		}
	}

	req := sandboxer.LaunchSandboxRequest{
		JobId:             e.jobId,
		JobEnv:            e.rCtx.env,
		JobContainer:      jobContainer,
		ServiceContainers: serviceContainers,
	}
	if res, err := runtime.LaunchSandbox(ctx, req); err != nil {
		return err
	} else {
		e.rCtx.sandbox = res.Sandbox
	}
	return nil
}

func (e *JobExecutor) Run(ctx context.Context) error {
	return e.rCtx.RunJob(ctx)
}

func (e *JobExecutor) Finalize(ctx context.Context, runtime sandboxer.SandboxRuntime) error {
	req := sandboxer.TerminateSandboxRequest{
		Sandbox: e.rCtx.sandbox,
	}
	_, err := runtime.TerminateSandbox(ctx, req)
	return err
}

func toContainerConfig(ctx context.Context, rCtx *JobRunContext, container *workflows.Container) (*container.ContainerConfig, error) {
	return nil, nil
}

type JobRun struct {
	UUID string
	ID   string
	Name string

	Container workflows.Evaluable[*workflows.Container]
	Services  workflows.Evaluable[map[string]*workflows.Container]

	Env      workflows.Evaluable[map[string]string]
	Steps    []StepRun
	Outputs  workflows.Evaluable[map[string]string]
	Defaults workflows.Evaluable[workflows.Defaults]
	// Environment
}
