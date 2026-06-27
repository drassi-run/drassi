/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package reporter

import (
	"context"
	"fmt"
	"time"

	runnerv1 "code.gitea.io/actions-proto-go/runner/v1"
	"code.gitea.io/actions-proto-go/runner/v1/runnerv1connect"
	"connectrpc.com/connect"
	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/util/context"
	"drassi.run/core/util/reactive"
	"github.com/chainguard-dev/clog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var resultMap = map[records.Result]runnerv1.Result{
	"":                      runnerv1.Result_RESULT_UNSPECIFIED,
	records.ResultSuccess:   runnerv1.Result_RESULT_SUCCESS,
	records.ResultFailure:   runnerv1.Result_RESULT_FAILURE,
	records.ResultCancelled: runnerv1.Result_RESULT_CANCELLED,
	records.ResultSkipped:   runnerv1.Result_RESULT_SKIPPED,
}

var ErrServerCancel = fmt.Errorf("task cancelled by Gitea server")

func New(
	taskId int64,
	client runnerv1connect.RunnerServiceClient,
	contextual xcontext.Provider,
	logStreamer *LogStreamer,
	cancel context.CancelCauseFunc,
) *Reporter {
	r := &Reporter{
		taskId:      taskId,
		client:      client,
		contextual:  contextual,
		logStreamer: logStreamer,
		stopCh:      make(chan struct{}),
		cancel:      cancel,
		ws:          reactive.NewWaitState(reactive.StateCreated),
	}

	return r
}

type Reporter struct {
	taskId      int64
	client      runnerv1connect.RunnerServiceClient
	contextual  xcontext.Provider
	logStreamer *LogStreamer

	timer  *time.Timer
	stopCh chan struct{}
	cancel context.CancelCauseFunc // hook that triggered by server used to cancel running job
	ws     *reactive.WaitState[reactive.State]

	jobUid     string
	jobState   *runnerv1.TaskState
	jobOutputs map[string]string
	stepStates map[string]*runnerv1.StepState
}

func (r *Reporter) init(spec *executor.JobSpec) {
	r.jobUid = spec.Uid
	r.jobOutputs = make(map[string]string)
	r.jobState = &runnerv1.TaskState{
		Id:    r.taskId,
		Steps: make([]*runnerv1.StepState, len(spec.Steps)),
	}
	r.stepStates = make(map[string]*runnerv1.StepState, len(spec.Steps))
	for i, step := range spec.Steps {
		s := &runnerv1.StepState{Id: int64(i)}
		r.jobState.Steps[i] = s
		r.stepStates[step.Id] = s
	}
}

func (r *Reporter) DecorateJobRun(task *executor.JobTask) executor.JobRun {
	if task.Stage == executor.StagePre {
		r.init(task.JobSpec())
	}

	run := task.Run
	if task.Stage != executor.StageMain {
		return run
	}

	return func(ctx context.Context) (*records.JobResult, error) {
		r.jobState.StartedAt = timestamppb.Now()
		r.flush()

		rec, err := run(ctx)

		r.jobState.StoppedAt = timestamppb.Now()
		r.jobState.Result = resultMap[rec.Result]
		r.jobOutputs = rec.Outputs
		r.flush()

		return rec, err
	}
}

func (r *Reporter) DecorateStepRun(task *executor.StepTask) executor.StepRun {
	if task.Stage != executor.StageMain {
		return task.Run
	}

	run := task.Run
	stepState := r.stepStates[task.StepId()]
	return func(ctx context.Context) (*records.StepResult, error) {
		stepState.StartedAt = timestamppb.Now()
		stepState.LogIndex = r.logStreamer.Offset()
		r.flush()

		rec, err := run(ctx)

		stepState.StoppedAt = timestamppb.Now()
		stepState.Result = resultMap[rec.Conclusion]
		stepState.LogLength = r.logStreamer.Offset() - stepState.LogIndex
		r.flush()

		return rec, err
	}
}

func (r *Reporter) flush() {
	r.timer.Reset(0)
}

func (r *Reporter) Start() error {
	if r.ws.Get() != reactive.StateCreated {
		return fmt.Errorf("reporter already started")
	}

	r.timer = time.NewTimer(time.Second)
	go r.start()
	return nil
}

func (r *Reporter) start() {
	r.ws.Set(reactive.StateRunning)
	defer r.ws.Set(reactive.StateStopped)

	for {
		select {
		case <-r.timer.C:
			r.updateTask()
			r.timer.Reset(time.Second)
		case <-r.stopCh:
			r.updateTask()
			r.timer.Stop()
			return
		}
	}
}

func (r *Reporter) Close() error {
	close(r.stopCh)
	r.ws.Wait(reactive.StateStopped)

	if len(r.jobOutputs) > 0 {
		return fmt.Errorf("there are still outputs that have not been sent")
	}
	return nil
}

func (r *Reporter) updateTask() {
	ctx := r.contextual.Context()

	req := &runnerv1.UpdateTaskRequest{
		State:   r.jobState,
		Outputs: r.jobOutputs,
	}
	resp, err := r.client.UpdateTask(ctx, connect.NewRequest(req))
	if err != nil {
		clog.ErrorContextf(ctx, "failed to update Task: %v", err)
		return
	}

	msg := resp.Msg
	if msg.State != nil && msg.State.Result == runnerv1.Result_RESULT_CANCELLED {
		clog.ErrorContextf(ctx, "task %d cancelled by Gitea server", r.taskId)
		r.cancel(ErrServerCancel)
		return
	}

	// https://github.com/go-gitea/gitea/blob/v1.23.7/routers/api/actions/runner/runner.go#L174
	if len(r.jobOutputs) > 0 {
		for _, k := range msg.SentOutputs {
			delete(r.jobOutputs, k)
		}
	}
}
