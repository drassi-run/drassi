package executor

import (
	"context"

	"github.com/dungdm93/drasi/pkg/model/workflows"
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
	rCtx   *JobRunContext
	job    workflows.Job
	runner StepsRunner
}

func (s *JobRunner) Initialize(ctx context.Context) error {
	//TODO launch sandbox
	panic("implement me")
}

func (s *JobRunner) Run(ctx context.Context) error {
	//TODO run pre -> main -> post
	panic("implement me")
}

func (s *JobRunner) Finalize(ctx context.Context) error {
	//TODO terminate sandbox
	panic("implement me")
}
