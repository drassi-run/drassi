/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package log

import (
	"sync"
	"time"
)

type Batch interface {
	Empty() bool
	Lines() int
	Scan() ([]string, error)
}

type Batcher interface {
	Channel() <-chan Batch
	Update(u *Update)
	Close() error
}

func NewBatcher(threshold int, interval time.Duration) Batcher {
	timer := time.NewTimer(interval)
	timer.Stop()

	br := &batcher{
		threshold: threshold,
		interval:  interval,
		ch:        make(chan Batch, 3),
		stopCh:    make(chan struct{}),
		timer:     timer,
	}
	br.wg.Go(br.Run)
	return br
}

// batcher emit a Batch as soon as accumulated lines reach threshold
// or interval time has elapsed since first Update.
type batcher struct {
	threshold int
	interval  time.Duration

	wg     sync.WaitGroup
	mu     sync.Mutex
	ch     chan Batch
	stopCh chan struct{}
	timer  *time.Timer

	batch   sections // stating batch
	total   int      // pre-computed line count
	section *section // stating section
}

func (br *batcher) Channel() <-chan Batch {
	return br.ch
}

func (br *batcher) Update(u *Update) {
	br.mu.Lock()
	defer br.mu.Unlock()

	if br.batch.Empty() && br.section.Empty() {
		// first line of batch, start timer
		br.timer.Reset(br.interval)
	}

	if br.section == nil {
		br.section = newSection(u)
	} else if br.section.filePath != u.File {
		s := newSection(u)
		br.stageSection(s)
	} else {
		br.section.update(u)
	}

	if u.Complete {
		br.stageSection(nil)
	}

	if br.total+br.section.Lines() >= br.threshold {
		br.timer.Reset(0) // trigger emit immediately
	}
}

func (br *batcher) Run() {
	for {
		select {
		case <-br.timer.C:
			if b := br.flush(); !b.Empty() {
				br.ch <- b
			}
		case <-br.stopCh:
			br.timer.Stop()
			if b := br.flush(); !b.Empty() {
				br.ch <- b
			}
			return
		}
	}
}

// append current section to the sections (if any) and assign it to new one
func (br *batcher) stageSection(new *section) {
	if !br.section.Empty() {
		br.batch = append(br.batch, br.section)
		br.total += br.section.Lines()
	}
	br.section = new
}

// flush state and return the current batch if any.
// NOTE: batch is not send to ch here to avoid ch and mu block each other.
// (block mu will affect Update func)
func (br *batcher) flush() sections {
	br.mu.Lock()
	defer br.mu.Unlock()

	if br.section != nil {
		s := br.section.next()
		br.stageSection(s)
	}
	b := br.batch
	br.batch, br.total = nil, 0

	return b
}

func (br *batcher) Close() error {
	close(br.stopCh)
	br.wg.Wait() // waiting for last Batch sent
	close(br.ch)
	return nil
}
