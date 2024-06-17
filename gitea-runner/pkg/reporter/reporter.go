package reporter

import (
	"context"
	"fmt"
	"io"
	"sync"

	"code.gitea.io/actions-proto-go/runner/v1"
	"code.gitea.io/actions-proto-go/runner/v1/runnerv1connect"
	"connectrpc.com/connect"
	"github.com/dungdm93/drassi/core/pkg/executor/reporter"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type giteaReporter struct {
	ctx    context.Context
	taskId int64
	client runnerv1connect.RunnerServiceClient

	out io.Writer
	err io.Writer

	logOffset int64
	logRows   []*runnerv1.LogRow
	mu        sync.RWMutex
}

func New(ctx context.Context, taskId int64, client runnerv1connect.RunnerServiceClient) reporter.Reporter {
	r := &giteaReporter{
		ctx:    ctx,
		taskId: taskId,
		client: client,
	}

	r.out = reporter.NewLineWriter(r.appendLogLine)
	r.err = reporter.NewLineWriter(r.appendLogLine)
	return r
}

func (r *giteaReporter) Stdin() io.Reader {
	return nil
}

func (r *giteaReporter) Stdout() io.Writer {
	return r.out
}

func (r *giteaReporter) Stderr() io.Writer {
	return r.err
}

func (r *giteaReporter) appendLogLine(line string) error {
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

func (r *giteaReporter) uploadLog(noMore bool) error {
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

func (r *giteaReporter) AddIssue(issue *reporter.Issue) error {
	//TODO implement me
	panic("implement me")
}

func (r *giteaReporter) AttachFile(kind, name string, reader io.Reader) error {
	//TODO implement me
	panic("implement me")
}

func (r *giteaReporter) Close() error {
	if err := r.uploadLog(true); err != nil {
		return err
	}
	return nil
}
