package log

import "sync"

func NewChunker(softLimit int64) *Chunker {
	return &Chunker{
		softLimit: softLimit,
		ch:        make(chan Chunk, 100),
	}
}

// Chunker divide files into chunks when its size reach softLimit.
// A chunk can also consist of multiple sections from multiple files.
type Chunker struct {
	softLimit int64

	mu sync.Mutex
	ch chan Chunk

	chunk   Chunk    // current Chunk
	size    int64    // pre-computed Chunk size
	section *Section // current Section
}

func (cr *Chunker) Channel() <-chan Chunk {
	return cr.ch
}

func (cr *Chunker) Update(u *Update) {
	if c := cr.update(u); !c.Empty() {
		cr.ch <- c
	}
}

func (cr *Chunker) update(u *Update) Chunk {
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

// append current Section to the Chunk (if any) and assign it to new one
func (cr *Chunker) stageSection(new *Section) {
	if !cr.section.Empty() {
		cr.chunk = append(cr.chunk, cr.section)
		cr.size += cr.section.Size()
	}
	cr.section = new
}

func (cr *Chunker) Close() error {
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

// flush state and return the current Chunk if any.
// NOTE: The caller MUST be inside mu.Lock
func (cr *Chunker) flush() Chunk {
	c := cr.chunk
	cr.chunk, cr.size = nil, 0
	return c
}
