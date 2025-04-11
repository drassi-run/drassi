package reporter

import (
	"context"
	"io"
	"os"
	"sync"

	"drassi.run/core/pkg/executor"
	"drassi.run/gha-runner/pkg/chunk"
)

type FileTransporter interface {
	Transport(ctx context.Context, sr executor.StepRun, file string, waiter chunk.Waiter) (done func() error, err error)
	WaitAllComplete()
}

const chunkSize = 2 * 1024 * 1024 // 2MiB

type stepLogs2ResultService struct {
	svc *resultService
	wg  sync.WaitGroup
}

func (ft *stepLogs2ResultService) Transport(ctx context.Context, sr executor.StepRun, file string, waiter chunk.Waiter) (done func() error, err error) {
	handle := func(ctx context.Context, r io.Reader, i int64) error {
		return ft.svc.UploadStepLogs(ctx, sr, r)
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	ps := chunk.NewChunker(f,
		chunk.WithSoftLimit(chunkSize),
		chunk.WithLineSafety(true),
		chunk.WithFollowInput(true),
		chunk.WithWaiter(waiter),
	)

	// TODO
	if err = ps.Run(ctx, handle); err != nil {
		return nil, err
	}

	done = func() error {
		ps.Complete()
		return nil
	}
	return
}

func (ft *stepLogs2ResultService) WaitAllComplete() {
	ft.wg.Wait()
}

type stepLogs2TaskService struct {
	svc *taskService
	wg  sync.WaitGroup
}

func (ft *stepLogs2TaskService) Transport(ctx context.Context, _ executor.StepRun, file string, _ chunk.Waiter) (done func() error, err error) {
	recordId := "" // TODO

	done = func() error {
		return ft.upload(ctx, recordId, file)
	}
	return
}

func (ft *stepLogs2TaskService) upload(ctx context.Context, recordId, file string) error {
	ft.wg.Add(1)
	defer ft.wg.Done()

	if f, err := os.Open(file); err != nil {
		return err
	} else {
		return ft.svc.UploadLog(ctx, recordId, f)
	}
}

func (ft *stepLogs2TaskService) WaitAllComplete() {
	ft.wg.Wait()
}
