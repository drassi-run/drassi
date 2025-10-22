package log

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/streaming"
)

type Update struct {
	File     string // File location
	Complete bool   // Is File completed or not
	Offset   int64  // Offset at the end of log record
	Line     int    // Line number of log record
}

func newSection(u *Update) *Section {
	return &Section{
		filePath:  u.File,
		endOffset: u.Offset,
		endLine:   u.Line,
	}
}

// Section represents a segment - contiguous region of a file.
//
// It describes a logical slice of filePath bounded by byte offsets
// [startOffset, endOffset), and the corresponding 0-indexed line range
// [startLine, endLine).
type Section struct {
	filePath    string
	startOffset int64
	endOffset   int64
	startLine   int
	endLine     int
	eof         bool
}

func (s *Section) Empty() bool {
	return s == nil || s.Size() <= 0
}

func (s *Section) Size() int64 {
	return s.endOffset - s.startOffset
}

func (s *Section) Lines() int {
	return s.endLine - s.startLine
}

// EOF indicate Section is end of file or not
func (s *Section) EOF() bool {
	return s.eof
}

// update endOffset and endLine from u
func (s *Section) update(u *Update) *Section {
	s.endOffset = u.Offset
	s.endLine = u.Line
	return s
}

// next return new Section start from the end of current one
func (s *Section) next() *Section {
	return &Section{
		filePath:    s.filePath,
		startOffset: s.endOffset,
		endOffset:   s.endOffset,
		startLine:   s.endLine,
		endLine:     s.endLine,
	}
}

func (s *Section) Reader() (io.ReadSeekCloser, error) {
	f, err := os.Open(s.filePath)
	if err != nil {
		return nil, err
	}

	var size int64
	if stat, err := f.Stat(); err != nil {
		return nil, err
	} else {
		size = stat.Size()
	}

	s.eof = s.endOffset >= size
	if s.startOffset <= 0 && size <= s.endOffset {
		return f, nil
	}
	sr := io.NewSectionReader(f, s.startOffset, s.Size())
	return rsc{sr, f}, nil
}

// Chunk aggregates multiple Section from different files and treat as a single logical block.
type Chunk []*Section

func (c Chunk) Empty() bool {
	for _, s := range c {
		if !s.Empty() {
			return false
		}
	}
	return true
}

func (c Chunk) Size() int64 {
	t := int64(0)
	for _, s := range c {
		t += s.Size()
	}
	return t
}

func (c Chunk) Lines() int {
	t := 0
	for _, s := range c {
		t += s.Lines()
	}
	return t
}

func (c Chunk) Reader() (io.ReadSeekCloser, error) {
	switch len(c) {
	case 0:
		return empty, nil
	case 1:
		return c[0].Reader()
	default:
		var m = new(multiReader)
		for _, r := range c {
			reader, err := r.Reader()
			if err != nil {
				_ = m.Close()
				return nil, err
			}
			m.readers = append(m.readers, reader)
			m.sizes = append(m.sizes, r.Size())
		}
		return m, nil
	}
}

var empty = streaming.NopCloser(strings.NewReader(""))

type rsc struct {
	io.ReadSeeker
	io.Closer
}

type multiReader struct {
	readers []io.ReadSeekCloser
	sizes   []int64
	idx     int   // current reader index
	off     int64 // current global offset
}

func (m *multiReader) Read(p []byte) (n int, err error) {
	for m.idx < len(m.readers) && n < len(p) {
		rn, rerr := m.readers[m.idx].Read(p[n:])
		n += rn
		m.off += int64(rn)
		if rerr == nil {
			continue // buffer is full or last bytes in reader
		}
		if rerr != io.EOF {
			return n, rerr
		}
		m.idx++ // current reader EOF, move to next one
	}
	if n > 0 {
		return n, nil
	}
	return 0, io.EOF
}

func (m *multiReader) Seek(offset int64, whence int) (int64, error) {
	var abs int64

	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = m.off + offset
	case io.SeekEnd:
		for _, s := range m.sizes {
			abs += s
		}
		abs += offset
	default:
		return 0, errors.New("invalid whence")
	}

	if abs < 0 {
		return 0, errors.New("negative offset")
	}

	// Find reader where seeking to
	var sum int64
	var idx int
	for i, size := range m.sizes {
		if abs < sum+size || i+1 == len(m.sizes) {
			idx = i
			break
		}
		sum += size
	}

	// perform seeking
	if _, err := m.readers[idx].Seek(abs-sum, io.SeekStart); err != nil {
		return 0, err
	}
	m.idx, m.off = idx, abs

	// Reset subsequent readers to start
	for i := idx + 1; i < len(m.readers); i++ {
		m.readers[i].Seek(0, io.SeekStart)
	}

	return abs, nil
}

func (m *multiReader) Close() error {
	errs := make([]error, len(m.readers))
	for i, r := range m.readers {
		errs[i] = r.Close()
	}
	return errors.Join(errs...)
}
