package service

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"code.gitea.io/actions-proto-go/runner/v1"
	"code.gitea.io/actions-proto-go/runner/v1/runnerv1connect"
	"connectrpc.com/connect"
	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/logging"
	"drassi.run/core/pkg/executor/reporter"
	"drassi.run/core/pkg/model/dossiers"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var resultMap = map[dossiers.Result]runnerv1.Result{
	"":                       runnerv1.Result_RESULT_UNSPECIFIED,
	dossiers.ResultSuccess:   runnerv1.Result_RESULT_SUCCESS,
	dossiers.ResultFailure:   runnerv1.Result_RESULT_FAILURE,
	dossiers.ResultCancelled: runnerv1.Result_RESULT_CANCELLED,
	dossiers.ResultSkipped:   runnerv1.Result_RESULT_SKIPPED,
}

type GiteaReporter struct {
	ctx    context.Context
	taskId int64
	client runnerv1connect.RunnerServiceClient

	out io.Writer
	err io.Writer

	logOffset   int64
	logRows     []*runnerv1.LogRow
	taskOutputs map[string]string
	taskState   *runnerv1.TaskState
	stepStates  map[string]*runnerv1.StepState
	mu          sync.RWMutex
}

func NewReporter(
	ctx context.Context,
	taskId int64,
	jobRun *executor.JobRun,
	client runnerv1connect.RunnerServiceClient,
) *GiteaReporter {
	r := &GiteaReporter{
		ctx:    ctx,
		taskId: taskId,
		client: client,
	}

	r.out = logging.NewLineWriter(r.appendLogLine)
	r.err = logging.NewLineWriter(r.appendLogLine)

	r.taskOutputs = make(map[string]string)
	stepStates := make([]*runnerv1.StepState, len(jobRun.Steps))
	r.stepStates = make(map[string]*runnerv1.StepState, len(jobRun.Steps))
	for i, step := range jobRun.Steps {
		s := &runnerv1.StepState{Id: int64(i)}
		stepStates[i] = s
		r.stepStates[step.StepId()] = s
	}
	r.taskState = &runnerv1.TaskState{
		Id:    taskId,
		Steps: stepStates,
	}

	return r
}

func (r *GiteaReporter) Stdin() io.Reader {
	return nil
}

func (r *GiteaReporter) Stdout() io.Writer {
	return r.out
}

func (r *GiteaReporter) Stderr() io.Writer {
	return r.err
}

func (r *GiteaReporter) appendLogLine(line string) error {
	line = strings.TrimRight(line, "\r\n")

	row := &runnerv1.LogRow{
		Time:    timestamppb.Now(),
		Content: line,
	}

	r.logRows = append(r.logRows, row)
	if len(r.logRows) >= 50 {
		_ = r.uploadLog(false)
	}
	return nil
}

func (r *GiteaReporter) uploadLog(noMore bool) error {
	req := &runnerv1.UpdateLogRequest{
		TaskId: r.taskId,
		Index:  r.logOffset,
		Rows:   r.logRows,
		NoMore: noMore,
	}
	resp, err := r.client.UpdateLog(r.ctx, connect.NewRequest(req))
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

func (r *GiteaReporter) StartJob() {
	r.taskState.StartedAt = timestamppb.Now()

	_ = r.updateTask()
}

func (r *GiteaReporter) EndJob(result dossiers.Result, outputs map[string]string) {
	r.taskState.StoppedAt = timestamppb.Now()
	r.taskState.Result = resultMap[result]

	for k, v := range outputs {
		r.taskOutputs[k] = v
	}

	_ = r.updateTask()
}

func (r *GiteaReporter) StartStep(stepId string) {
	stepState := r.stepStates[stepId]

	stepState.StartedAt = timestamppb.Now()
	stepState.LogIndex = r.logOffset + int64(len(r.logRows))

	_ = r.updateTask()
}

func (r *GiteaReporter) EndStep(stepId string, result dossiers.Result) {
	stepState := r.stepStates[stepId]

	stepState.StoppedAt = timestamppb.Now()
	stepState.Result = resultMap[result]
	stepState.LogLength = r.logOffset + int64(len(r.logRows)) - stepState.LogIndex

	_ = r.updateTask()
}

func (r *GiteaReporter) updateTask() error {
	req := &runnerv1.UpdateTaskRequest{
		State:   r.taskState,
		Outputs: r.taskOutputs,
	}
	resp, err := r.client.UpdateTask(r.ctx, connect.NewRequest(req))
	if err != nil {
		return err
	}

	// TODO: gitea server cancel job
	//if resp.Msg.State != nil && resp.Msg.State.Result == runnerv1.Result_RESULT_CANCELLED {
	//	r.cancel()
	//}
	for _, k := range resp.Msg.SentOutputs {
		delete(r.taskOutputs, k)
	}
	return nil
}

func (r *GiteaReporter) AddIssue(issue *reporter.Issue) error {
	//TODO implement me
	panic("implement me")
}

func (r *GiteaReporter) AttachFile(kind, name string, reader io.Reader) error {
	//TODO implement me
	panic("implement me")
}

func (r *GiteaReporter) Close() error {
	if err := r.uploadLog(true); err != nil {
		return err
	}

	if err := r.updateTask(); err != nil {
		return err
	}

	if len(r.taskOutputs) > 0 {
		return fmt.Errorf("there are still outputs that have not been sent")
	}
	return nil
}
