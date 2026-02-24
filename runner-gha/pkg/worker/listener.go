package worker

import (
	"context"
	"fmt"

	"drassi.run/core/pkg/executor"
	"drassi.run/gha-runner/pkg/types"
	"go.uber.org/dig"
)

type ghaListener struct {
	executor.NoopJobListener
	executor.NoopStepListener

	order   int
	records map[string]*types.Record
}

func (l *ghaListener) OnInitializeJob(exec executor.JobExecutor, scope *dig.Scope) executor.EventHandler {
	r := &types.Record{
		Id:  executor.JobId(exec),
		Uid: executor.JobUid(exec),
		Object: &types.JobObject{
			JobRun: exec.JobRun(),
		},
	}
	l.putRecord("job", r)

	return &jobInitEventHandler{exec, l}
}

func (l *ghaListener) OnRunJob(exec executor.JobExecutor) executor.EventHandler {
	//TODO implement me
	panic("implement me")
}

func (l *ghaListener) OnFinalizeJob(exec executor.JobExecutor) executor.EventHandler {
	//TODO implement me
	panic("implement me")
}

func (l *ghaListener) OnInitializeStep(exec executor.StepExecutor, scope *dig.Scope) executor.EventHandler {
	//TODO implement me
	panic("implement me")
}

func (l *ghaListener) OnRunStep(exec executor.StepExecutor, stage executor.Stage) executor.EventHandler {
	//TODO implement me
	panic("implement me")
}

func (l *ghaListener) putRecord(key string, r *types.Record) {
	l.order++
	r.Order = l.order
	l.records[key] = r
}

func (l *ghaListener) getRecord(key string) *types.Record {
	return l.records[key]
}

type jobInitEventHandler struct {
	exec     executor.JobExecutor
	listener *ghaListener
}

func (h *jobInitEventHandler) Begin(ctx context.Context) error {
	r := &types.Record{
		Id:  executor.JobId(h.exec),
		Uid: executor.JobUid(h.exec),
	}
	h.listener.putRecord("__init__", r)
	return nil
}

func (h *jobInitEventHandler) End(err error) error {
	r := &types.Record{
		Id:  executor.JobId(h.exec),
		Uid: executor.JobUid(h.exec),
	}
	h.listener.putRecord("__complete__", r)
	return err
}

type stepInitEventHandler struct {
	exec     executor.StepExecutor
	listener *ghaListener
}

func (h *stepInitEventHandler) Begin(ctx context.Context) error {
	run := h.exec.StepRun()
	if run.PreTask() != nil {
		h.initStep(executor.StagePre, run)
	}
	h.initStep(executor.StageMain, run)
	if run.PostTask() != nil {
		h.initStep(executor.StagePost, run)
	}
	return nil
}

func (h *stepInitEventHandler) initStep(stage executor.Stage, run executor.StepRun) {
	r := &types.Record{
		Id:  run.StepId(),
		Uid: run.Base().Uid,
		Object: &types.StepObject{
			StepRun: run,
			Stage:   stage,
		},
	}
	h.listener.putRecord(fmt.Sprintf("step/%s/%s", stage, run.StepId()), r)
}

func (h *stepInitEventHandler) End(err error) error {
	return err
}
