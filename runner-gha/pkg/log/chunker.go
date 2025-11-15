/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package log

import (
	"io"
	"sync"
)

type Chunk interface {
	Empty() bool
	Size() int64
	Lines() int
	Reader() (io.ReadSeekCloser, error)
}

type Chunker interface {
	Channel() <-chan Chunk
	Update(u *Update)
	Close() error
}

func NewChunker(softLimit int64) Chunker {
	return &chunker{
		softLimit: softLimit,
		ch:        make(chan Chunk, 100),
	}
}

// chunker divide files into chunks when its size reach softLimit.
// A chunk can also consist of multiple sections from multiple files.
type chunker struct {
	softLimit int64

	mu sync.Mutex
	ch chan Chunk

	chunk   chunk    // stating chunk
	size    int64    // pre-computed chunk size
	section *section // stating section
}

func (cr *chunker) Channel() <-chan Chunk {
	return cr.ch
}

func (cr *chunker) Update(u *Update) {
	if c := cr.update(u); !c.Empty() {
		cr.ch <- c
	}
}

func (cr *chunker) update(u *Update) chunk {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if cr.section == nil {
		cr.section = newSection(u)
	} else if cr.section.filePath != u.File {
		s := newSection(u)
		cr.stageSection(s)
	} else {
		cr.section.update(u)
	}

	if u.Complete {
		cr.stageSection(nil)
	} else if cr.size+cr.section.Size() >= cr.softLimit {
		s := cr.section.next()
		cr.stageSection(s)
	}

	if cr.size >= cr.softLimit {
		return cr.flush()
	}
	return nil
}

// append current section to the chunk (if any) and assign it to new one
func (cr *chunker) stageSection(new *section) {
	if !cr.section.Empty() {
		cr.chunk = append(cr.chunk, cr.section)
		cr.size += cr.section.Size()
	}
	cr.section = new
}

func (cr *chunker) Close() error {
	cr.mu.Lock()
	cr.stageSection(nil)
	c := cr.flush()
	cr.mu.Unlock()

	if !c.Empty() {
		cr.ch <- c
	}
	close(cr.ch)
	return nil
}

// flush state and return the current chunk if any.
// NOTE: The caller MUST be inside mu.Lock
func (cr *chunker) flush() chunk {
	c := cr.chunk
	cr.chunk, cr.size = nil, 0
	return c
}
