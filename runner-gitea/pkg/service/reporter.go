/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package service

import (
	"context"
	"fmt"
	"io"
	"maps"
	"strings"
	"sync"

	runnerv1 "code.gitea.io/actions-proto-go/runner/v1"
	"code.gitea.io/actions-proto-go/runner/v1/runnerv1connect"
	"connectrpc.com/connect"
	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/reporter"
	"drassi.run/core/pkg/model/records"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var resultMap = map[records.Result]runnerv1.Result{
	"":                      runnerv1.Result_RESULT_UNSPECIFIED,
	records.ResultSuccess:   runnerv1.Result_RESULT_SUCCESS,
	records.ResultFailure:   runnerv1.Result_RESULT_FAILURE,
	records.ResultCancelled: runnerv1.Result_RESULT_CANCELLED,
	records.ResultSkipped:   runnerv1.Result_RESULT_SKIPPED,
}

type GiteaReporter struct {
	ctx    context.Context
	taskId int64
	jobUid string
	client runnerv1connect.RunnerServiceClient

	logOffset  int64
	logRows    []*runnerv1.LogRow
	jobOutputs map[string]string
	jobState   *runnerv1.TaskState
	stepStates map[string]*runnerv1.StepState
	mu         sync.RWMutex
}

func NewReporter(
	ctx context.Context,
	taskId int64,
	client runnerv1connect.RunnerServiceClient,
) *GiteaReporter {
	r := &GiteaReporter{
		ctx:        ctx,
		taskId:     taskId,
		client:     client,
		logOffset:  0,
		logRows:    make([]*runnerv1.LogRow, 0),
		jobOutputs: make(map[string]string),
	}

	return r
}

func (r *GiteaReporter) StartJob(ctx context.Context, je executor.JobExecutor) error {
	if r.jobUid != "" {
		if r.jobUid == executor.JobUid(je) {
			return fmt.Errorf("job already running")
		} else {
			return fmt.Errorf("another job is running")
		}
	}

	r.jobUid = executor.JobUid(je)
	jobRun := je.JobRun()

	r.jobState = &runnerv1.TaskState{
		Id:        r.taskId,
		StartedAt: timestamppb.Now(),
		Steps:     make([]*runnerv1.StepState, len(jobRun.Steps)),
	}

	r.stepStates = make(map[string]*runnerv1.StepState, len(jobRun.Steps))
	for i, step := range jobRun.Steps {
		s := &runnerv1.StepState{Id: int64(i)}
		r.jobState.Steps[i] = s
		r.stepStates[step.StepId()] = s
	}

	return r.updateTask(ctx)
}

func (r *GiteaReporter) EndJob(ctx context.Context, je executor.JobExecutor, result *records.Job) error {
	if r.jobUid == "" {
		return fmt.Errorf("no job already running")
	}
	if r.jobUid != executor.JobUid(je) {
		return fmt.Errorf("another job is running")
	}

	r.jobState.StoppedAt = timestamppb.Now()
	r.jobState.Result = resultMap[result.Result]

	maps.Copy(r.jobOutputs, result.Outputs)
	return r.updateTask(ctx)
}

func (r *GiteaReporter) StartStep(ctx context.Context, stage executor.Stage, se executor.StepExecutor) error {
	if stage != executor.StageMain {
		// Gitea only report main stage for now
		return nil
	}
	if se.ParentExecutor() != nil {
		// ignore report embeded step
		return nil
	}

	stepState := r.stepStates[executor.StepId(se)]
	stepState.StartedAt = timestamppb.Now()
	stepState.LogIndex = r.logOffset + int64(len(r.logRows))

	return r.updateTask(ctx)
}

func (r *GiteaReporter) EndStep(ctx context.Context, stage executor.Stage, se executor.StepExecutor, result *records.Step) error {
	if stage != executor.StageMain {
		// Gitea only report main stage for now
		return nil
	}
	if se.ParentExecutor() != nil {
		// ignore report embeded step
		return nil
	}

	stepState := r.stepStates[executor.StepId(se)]

	stepState.StoppedAt = timestamppb.Now()
	stepState.Result = resultMap[result.Conclusion]
	stepState.LogLength = r.logOffset + int64(len(r.logRows)) - stepState.LogIndex

	return r.updateTask(ctx)
}

func (r *GiteaReporter) AddIssue(ctx context.Context, issue *reporter.Issue) error {
	//TODO implement me
	panic("implement me")
}

func (r *GiteaReporter) AttachFile(kind, name string, reader io.Reader) error {
	//TODO implement me
	panic("implement me")
}

func (r *GiteaReporter) Close() error {
	if err := r.uploadLog(r.ctx, true); err != nil {
		return err
	}

	if err := r.updateTask(r.ctx); err != nil {
		return err
	}

	if len(r.jobOutputs) > 0 {
		return fmt.Errorf("there are still outputs that have not been sent")
	}
	return nil
}

func (r *GiteaReporter) Log(ctx context.Context, msg string) error {
	msg = strings.TrimRight(msg, "\r\n")

	row := &runnerv1.LogRow{
		Time:    timestamppb.Now(),
		Content: msg,
	}

	r.logRows = append(r.logRows, row)
	if len(r.logRows) >= 50 {
		_ = r.uploadLog(ctx, false)
	}
	return nil
}

func (r *GiteaReporter) uploadLog(ctx context.Context, noMore bool) error {
	req := &runnerv1.UpdateLogRequest{
		TaskId: r.taskId,
		Index:  r.logOffset,
		Rows:   r.logRows,
		NoMore: noMore,
	}
	resp, err := r.client.UpdateLog(ctx, connect.NewRequest(req))
	if err != nil {
		return err
	}

	ack := resp.Msg.AckIndex
	if ack < r.logOffset {
		return fmt.Errorf("submitted logs are lost")
	}

	r.mu.Lock()
	r.logRows = r.logRows[ack-r.logOffset:]
	r.logOffset = ack
	r.mu.Unlock()

	if noMore && ack < r.logOffset+int64(len(r.logRows)) {
		return fmt.Errorf("not all logs are submitted")
	}

	return nil
}

func (r *GiteaReporter) updateTask(ctx context.Context) error {
	req := &runnerv1.UpdateTaskRequest{
		State:   r.jobState,
		Outputs: r.jobOutputs,
	}
	resp, err := r.client.UpdateTask(ctx, connect.NewRequest(req))
	if err != nil {
		return err
	}

	// TODO: gitea server cancel job
	//if resp.Msg.State != nil && resp.Msg.State.Result == runnerv1.Result_RESULT_CANCELLED {
	//	r.cancel()
	//}
	for _, k := range resp.Msg.SentOutputs {
		delete(r.jobOutputs, k)
	}
	return nil
}
