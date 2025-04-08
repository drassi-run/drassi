package reporter

import (
	"context"
	"drassi.run/core/pkg/executor"
	"drassi.run/gha-runner/pkg/chunk"
	"io"
	"os"
)

type FileHandler func(ctx context.Context, file string) (done func() error, err error)

const chunkSize = 2 * 1024 * 1024 // 2MiB

func ResultServiceStepLogsUploader(svc *resultService, sr executor.StepRun) FileHandler {
	handle := func(ctx context.Context, r io.Reader, i int64) error {
		return svc.UploadStepLogs(ctx, sr, r)
	}

	return func(ctx context.Context, file string) (done func() error, err error) {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}

		ps := chunk.NewProcessor(f,
			chunk.WithSoftChunkSize(chunkSize),
			chunk.WithLineSafety(true),
			chunk.WithFollowInput(true),
		)

		// TODO
		ps.Run(ctx, handle)

		done = func() error {
			ps.Complete()
			return nil
		}
		return done, nil
	}
}

func TaskServiceStepLogsUploader(svc *taskService, recordId string) FileHandler {
	return func(ctx context.Context, file string) (done func() error, err error) {
		done = func() error {
			if f, err := os.Open(file); err != nil {
				return err
			} else {
				return svc.UploadLog(ctx, recordId, f)
			}
		}
		return done, nil
	}
}
