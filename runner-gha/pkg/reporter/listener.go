package reporter

import (
	"sync"

	"drassi.run/core/pkg/executor"
	xcontext "drassi.run/core/util/context"
	xotel "drassi.run/core/util/otel"
	"drassi.run/gha-runner/pkg/reporter/log"
	"drassi.run/gha-runner/pkg/reporter/service"
)

type Listener interface {
	OnNewStep(sr executor.StepRun)
	OnLogRecord(stepUid string, update *log.Update)
	OnCompleteStep(sr executor.StepRun)
}

type resultStepLogListener struct {
	svc      *service.ResultService
	ctx      xcontext.Provider
	chunkers map[string]*log.Chunker
	mu       sync.Mutex
	wg       sync.WaitGroup
}

func (l *resultStepLogListener) logChunker(stepId string) *log.Chunker {
	l.mu.Lock()
	defer l.mu.Unlock()

	if c, ok := l.chunkers[stepId]; ok {
		return c
	}

	c := log.NewChunker(1000) // TODO
	l.chunkers[stepId] = c
	l.start(stepId, c)
	return c
}

func (l *resultStepLogListener) start(stepId string, chunker *log.Chunker) {
	l.wg.Add(1)
	go func() {
		defer l.wg.Done()

		ctx, logger := xotel.ChildLogger(
			l.ctx.Context(),
			xotel.ToSlogAttrs(xotel.DrassiStep(stepId)),
		)

		for c := range chunker.Channel() {
			if err := l.upload(ctx, c); err != nil {
				logger.Error("error while upload log chunk to result server", err)
			}
		}
	}()
}
