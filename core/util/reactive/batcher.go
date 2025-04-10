/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package reactive

import (
	"fmt"
	"sync"
	"time"
)

type Batcher[E any] interface {
	Put(item E) error
	Start(fn func([]E)) error
	Stop()
}

type State int

const (
	StateCreated State = iota
	StateRunning
	StateStopped
)

const defaultCap = 5

func NewThrottleBatcher[E any](softLimit int, delay time.Duration) Batcher[E] {
	timer := time.NewTimer(0)
	timer.Stop()
	ws := NewWaitState(StateCreated)

	return &throttleBatcher[E]{
		queue:  make([]E, 0, defaultCap),
		limit:  softLimit,
		delay:  delay,
		timer:  timer,
		stopCh: make(chan struct{}),
		ws:     ws,
	}
}

type throttleBatcher[E any] struct {
	queue []E
	limit int
	delay time.Duration

	timer *time.Timer
	mu    sync.Mutex

	stopCh chan struct{}
	ws     *WaitState[State]
}

func (b *throttleBatcher[E]) Put(item E) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.ws.Get() == StateStopped {
		return fmt.Errorf("batcher is stopped")
	}

	b.queue = append(b.queue, item)
	l := len(b.queue)
	if l == 1 {
		b.timer.Reset(b.delay)
	} else if l >= b.limit {
		b.timer.Reset(0)
	}
	return nil
}

func (b *throttleBatcher[E]) Start(fn func([]E)) error {
	if b.ws.Get() != StateCreated {
		return fmt.Errorf("batcher already started")
	}

	go b.start(fn)
	return nil
}

func (b *throttleBatcher[E]) start(fn func([]E)) {
	b.ws.Set(StateRunning)
	defer b.ws.Set(StateStopped)

	for {
		select {
		case <-b.timer.C:
			b.process(fn)
		case <-b.stopCh:
			b.timer.Stop()
			b.process(fn) // flush remaining items
			return
		}
	}
}

func (b *throttleBatcher[E]) process(fn func([]E)) {
	items := b.gather()
	if len(items) == 0 {
		return
	}

	fn(items)
}

func (b *throttleBatcher[E]) gather() (items []E) {
	b.mu.Lock()
	defer b.mu.Unlock()

	items = b.queue
	b.queue = make([]E, 0, defaultCap)
	return
}

func (b *throttleBatcher[E]) Stop() {
	close(b.stopCh)

	b.ws.Wait(StateStopped)
}
