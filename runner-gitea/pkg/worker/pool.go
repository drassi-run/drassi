package worker

import (
	"context"
	"fmt"
	"sync"

	"github.com/chainguard-dev/clog"
)

type Pool[T any] struct {
	con    int
	run    func(context.Context, T)
	avaiCh chan<- struct{}

	wg       sync.WaitGroup
	ctx      context.Context
	taskChan chan T
}

func NewPool[T any](concurrency int, available chan<- struct{}, run func(context.Context, T)) *Pool[T] {
	return &Pool[T]{
		con:    concurrency,
		run:    run,
		avaiCh: available,
	}
}

func (p *Pool[T]) worker(workerID int) {
	l := clog.FromContext(p.ctx)
	l.Infof("Worker %d: Started", workerID)

	p.avaiCh <- struct{}{}
	p.wg.Add(1)
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			l.Infof("Worker %d: Stopping", workerID)
		case task, ok := <-p.taskChan:
			if !ok {
				return // Channel closed, worker exits.
			}
			l.Infof("Worker %d: Processing task", workerID)
			p.run(p.ctx, task)
			p.avaiCh <- struct{}{} // Signal availability.
		}
	}
}

func (p *Pool[T]) Start(ctx context.Context) error {
	if p.ctx != nil {
		return fmt.Errorf("pool already started")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	p.ctx = ctx
	p.taskChan = make(chan T)

	for i := range p.con {
		go p.worker(i)
	}
	return nil
}

func (p *Pool[T]) Submit(task T) {
	p.taskChan <- task
}

func (p *Pool[T]) Close() error {
	close(p.taskChan)
	p.wg.Wait()
	return nil
}
