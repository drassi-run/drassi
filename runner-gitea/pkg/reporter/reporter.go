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

type Reporter struct {
	taskId      int64
	client      runnerv1connect.RunnerServiceClient
	contextual  xcontext.Provider
	logStreamer *LogStreamer

	timer  *time.Timer
	stopCh chan struct{}
	cancel context.CancelCauseFunc
	ws     *reactive.WaitState[reactive.State]

	jobUid     string
	jobState   *runnerv1.TaskState
	jobOutputs map[string]string
	stepStates map[string]*runnerv1.StepState
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

func (r *Reporter) StartJob(spec *executor.JobSpec) error {
	if r.jobUid != "" {
		if r.jobUid == spec.Uid {
			return fmt.Errorf("job already running")
		} else {
			return fmt.Errorf("another job is running")
		}
	}

	r.jobUid = spec.Uid

	r.jobState = &runnerv1.TaskState{
		Id:        r.taskId,
		StartedAt: timestamppb.Now(),
		Steps:     make([]*runnerv1.StepState, len(spec.Steps)),
	}
	r.jobOutputs = make(map[string]string)

	r.stepStates = make(map[string]*runnerv1.StepState, len(spec.Steps))
	for i, step := range spec.Steps {
		s := &runnerv1.StepState{Id: int64(i)}
		r.jobState.Steps[i] = s
		r.stepStates[step.StepId()] = s
	}

	r.timer.Reset(0)
	return nil
}

func (r *Reporter) EndJob(spec *executor.JobSpec, state *records.Job) error {
	if r.jobUid == "" {
		return fmt.Errorf("no job already running")
	}
	if r.jobUid != spec.Uid {
		return fmt.Errorf("another job is running")
	}

	r.jobState.StoppedAt = timestamppb.Now()
	r.jobState.Result = resultMap[state.Result]
	r.jobOutputs = state.Outputs

	r.timer.Reset(0)
	return nil
}

func (r *Reporter) StartStep(sr executor.StepSpec, stage executor.Stage) error {
	if stage != executor.StageMain {
		// Gitea only report main stage for now
		return nil
	}

	stepState := r.stepStates[sr.Base().StepId()]
	stepState.StartedAt = timestamppb.Now()
	stepState.LogIndex = r.logStreamer.Offset()

	r.timer.Reset(0)
	return nil
}

func (r *Reporter) EndStep(sr executor.StepSpec, stage executor.Stage, state *records.Step) error {
	if stage != executor.StageMain {
		// Gitea only report main stage for now
		return nil
	}

	stepState := r.stepStates[sr.Base().StepId()]

	stepState.StoppedAt = timestamppb.Now()
	stepState.Result = resultMap[state.Conclusion]
	stepState.LogLength = r.logStreamer.Offset() - stepState.LogIndex

	r.timer.Reset(0)
	return nil
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
