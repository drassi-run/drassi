package log

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/streaming"
)

type FileStatus int

const (
	FileOpen FileStatus = iota
	FileClose
	FileDeleted
)

type Update struct {
	file   string
	status FileStatus
	line   int64 // Total lines of the file
	size   int64 // Size of the file
}

type rsc struct {
	io.ReadSeeker
	io.Closer
}

var empty = streaming.NopCloser(strings.NewReader(""))

type multi struct {
	readers []io.ReadSeekCloser
	sizes   []int64
}

func (m *multi) Read(p []byte) (n int, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *multi) Seek(offset int64, whence int) (int64, error) {
	//TODO implement me
	panic("implement me")
}

type Section struct {
	filePath  string
	start     int64
	end       int64
	startLine int64
	endLine   int64
	eof       bool
}

func (s *Section) Size() int64 {
	return s.end - s.start
}

func (s *Section) Lines() int64 {
	return s.endLine - s.startLine
}

// EOF indicate Section is end of file or not
func (s *Section) EOF() bool {
	return s.eof
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

	s.eof = s.end >= size
	if s.start <= 0 && size <= s.end {
		return f, nil
	}
	sr := io.NewSectionReader(f, s.start, s.Size())
	return rsc{sr, f}, nil
}

func (m *multi) Close() error {
	errs := make([]error, len(m.readers))
	for i, r := range m.readers {
		errs[i] = r.Close()
	}
	return errors.Join(errs...)
}

type Chunk []*Section

func (c Chunk) Empty() bool {
	return len(c) == 0
}

func (c Chunk) Size() int64 {
	t := int64(0)
	for _, s := range c {
		t += s.Size()
	}
	return t
}

func (c Chunk) Lines() int64 {
	t := int64(0)
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
		var m = new(multi)
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
