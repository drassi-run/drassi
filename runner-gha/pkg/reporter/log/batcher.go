package log

import (
	"sync"
	"time"
)

type Batch []string

func NewBatcher(batchSize int64, throttleDelay time.Duration) *Batcher {
	timer := time.NewTimer(0)
	timer.Stop()

	return &Batcher{
		batchSize:     batchSize,
		throttleDelay: throttleDelay,
		timer:         timer,
		stopCh:        make(chan struct{}),
	}
}

type Batcher struct {
	batchSize     int64
	throttleDelay time.Duration

	mu     sync.Mutex
	timer  *time.Timer
	stopCh chan struct{}
}

func (br *Batcher) Update(u *Update) {
	//if c := br.update(u); c != nil {
	//	br.ch <- c
	//}

	if !br.mu.TryLock() {
		// There is a batch is processing
		return
	}
	defer br.mu.Unlock()

	l := int64(10) // TODO
	if l == 1 {
		br.timer.Reset(br.throttleDelay) // start timer
	} else if l >= br.batchSize {
		br.timer.Reset(0) // trigger process immediately
	}
}

func (br *Batcher) Run(fn func(Batch)) {
	for {
		select {
		case <-br.timer.C:
			br.process(fn)
		case <-br.stopCh:
			br.timer.Stop()
			br.process(fn) // flush remaining items
			return
		}
	}
}

func (br *Batcher) process(fn func(Batch)) {
	b := br.gather()
	if len(b) == 0 {
		return
	}

	fn(b)
}

func (br *Batcher) gather() Batch {
	br.mu.Lock()
	defer br.mu.Unlock()
	return nil
}

func (br *Batcher) Close() error {
	close(br.stopCh)
	return nil
}
