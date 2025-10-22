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

	mu       sync.Mutex
	ch       chan Chunk
	chunk    Chunk
	size     int64  // Total size of current chunk
	lastFile string // Last file processed
	offset   int64  // Offset (byte) of last file
	line     int    // Line number of last file
}

func (cr *Chunker) Channel() <-chan Chunk {
	return cr.ch
}

func (cr *Chunker) Update(u *Update) {
	if c := cr.update(u); c != nil {
		cr.ch <- c
	}
}

func (cr *Chunker) update(u *Update) Chunk {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if cr.lastFile != u.File {
		cr.lastFile = u.File
		cr.offset = 0
		cr.line = 0
	}

	if u.Complete {
		if cr.offset < u.Offset {
			cr.appendSection(u)
		}
		if cr.size >= cr.softLimit {
			return cr.flush()
		}
	} else {
		if cr.offset < u.Offset {
			cr.appendSection(u)
		}
		if cr.size >= cr.softLimit {
			return cr.flush()
		}
	}

	return nil
}

func (cr *Chunker) appendSection(u *Update) {
	s := &Section{
		filePath:    u.File,
		startOffset: cr.offset,
		endOffset:   u.Offset,
		startLine:   cr.line,
		endLine:     u.Line,
	}
	cr.chunk = append(cr.chunk, s)
	cr.size += u.Offset - cr.offset
	cr.offset, cr.line = u.Offset, u.Line
}

func (cr *Chunker) Close() error {
	cr.mu.Lock()
	c := cr.flush()
	cr.mu.Unlock()

	if c != nil {
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
