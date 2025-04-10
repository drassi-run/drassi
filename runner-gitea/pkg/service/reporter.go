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

	runnerv1 "code.gitea.io/actions-proto-go/runner/v1"
	"code.gitea.io/actions-proto-go/runner/v1/runnerv1connect"
	"connectrpc.com/connect"
	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/reporter"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/scribe"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/utils/set"
)

var resultMap = map[records.Result]runnerv1.Result{
	"":                      runnerv1.Result_RESULT_UNSPECIFIED,
	records.ResultSuccess:   runnerv1.Result_RESULT_SUCCESS,
	records.ResultFailure:   runnerv1.Result_RESULT_FAILURE,
	records.ResultCancelled: runnerv1.Result_RESULT_CANCELLED,
	records.ResultSkipped:   runnerv1.Result_RESULT_SKIPPED,
}

type GiteaReporter struct {
	taskId int64
	client runnerv1connect.RunnerServiceClient
	log    *LogStreamer

	jobUid     string
	jobState   *runnerv1.TaskState
	stepStates map[string]*runnerv1.StepState
}

func NewReporter(
	taskId int64,
	client runnerv1connect.RunnerServiceClient,
	logStreamer *LogStreamer,
) *GiteaReporter {
	r := &GiteaReporter{
		taskId: taskId,
		client: client,
		log:    logStreamer,
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

	return r.updateTask(ctx, nil)
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

	return r.updateTask(ctx, result.Outputs)
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
	stepState.LogIndex = r.log.Offset()

	return r.updateTask(ctx, nil)
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
	stepState.LogLength = r.log.Offset() - stepState.LogIndex

	return r.updateTask(ctx, nil)
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
	return nil
}

func (r *GiteaReporter) updateTask(ctx context.Context, output map[string]string) error {
	req := &runnerv1.UpdateTaskRequest{
		State:   r.jobState,
		Outputs: output,
	}
	resp, err := r.client.UpdateTask(ctx, connect.NewRequest(req))
	if err != nil || len(resp.Msg.SentOutputs) == 0 {
		return err
	}

	// https://github.com/go-gitea/gitea/blob/v1.23.7/routers/api/actions/runner/runner.go#L174
	s := scribe.FromContext(ctx)
	sentOutputs := set.New(resp.Msg.SentOutputs...)
	for k := range output {
		if !sentOutputs.Has(k) {
			s.Errorf("fail to update output %q to Gitea server", k)
		}
	}
	return nil
}
