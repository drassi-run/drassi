package transport

import (
	"context"
	"os"
	"sync"

	"github.com/chainguard-dev/clog"
)

func NewConveyor(courier []Courier) *Conveyor {
	c := &Conveyor{
		couriers: courier,
		queue:    make([]string, 0),
		complete: false,
		count:    make([]int, 0),
		doneCh:   make(chan int, 1),
	}
	c.cond = sync.NewCond(&c.mu)
	return c
}

type Conveyor struct {
	couriers []Courier
	queue    []string
	complete bool
	mu       sync.Mutex
	cond     *sync.Cond
	wg       sync.WaitGroup

	count  []int
	doneCh chan int
}

func (c *Conveyor) AddFile(f string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.queue = append(c.queue, f)
	c.count = append(c.count, 0)

	c.cond.Broadcast()
}

func (c *Conveyor) Complete() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.complete = true
	c.cond.Broadcast()
}

func (c *Conveyor) get(i int) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	if i < len(c.queue) {
		return c.queue[i]
	}
	if c.complete {
		return ""
	}

	// 3 cases can notify cond:
	// * new file added
	// * conveyor completed
	// * context is cancelled
	c.cond.Wait()

	if c.complete {
		return ""
	}
	if i < len(c.queue) {
		return c.queue[i]
	}
	return ""
}

func (c *Conveyor) done(i int) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.count[i]++
	if c.count[i] >= len(c.couriers) {
		return c.queue[i]
	}
	return ""
}

func (c *Conveyor) Start(ctx context.Context) {
	for _, t := range c.couriers {
		go c.run(ctx, t)
	}
	go c.cleanUp(ctx)
}

func (c *Conveyor) run(ctx context.Context, t Courier) {
	c.wg.Add(1)
	defer c.wg.Done()

	l := clog.FromContext(ctx)
	for i := 0; ; i++ {
		if i > 0 { // not first file
			prev := c.get(i - 1)
			if err := t.DoneFile(ctx, prev); err != nil {
				l.Errorf("failed to process file %s: %v", prev, err)
				return
			}
			c.doneCh <- i - 1
		}

		f := c.get(i)
		if f == "" { // no more files to process
			break
		}

		if err := t.NewFile(ctx, f); err != nil {
			l.Errorf("failed to process file %s: %v", f, err)
			return
		}
	}

	if err := t.Complete(ctx, 0); err != nil {
		l.Errorf("failed to complete transporter: %v", err)
	}
}

func (c *Conveyor) cleanUp(ctx context.Context) {
	l := clog.FromContext(ctx)

	// wakes up all waiting goroutines
	stop := context.AfterFunc(ctx, c.cond.Broadcast)
	defer stop()

	for {
		select {
		case idx, ok := <-c.doneCh:
			if !ok {
				return
			}
			if f := c.done(idx); f == "" {
				if err := os.Remove(f); err != nil {
					l.Errorf("failed to remove file %s: %v", f, err)
				}
			}
		case <-ctx.Done():
			close(c.doneCh)
			return
		}
	}
}

func (c *Conveyor) Wait() {
	c.wg.Wait()
	close(c.doneCh)
}
