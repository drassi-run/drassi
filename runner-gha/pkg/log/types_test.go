package log

import (
	"io"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/streaming"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type MultiReaderTestSuite struct {
	suite.Suite
	m *multiReader
}

func TestMultiTestSuite(t *testing.T) {
	suite.Run(t, new(MultiReaderTestSuite))
}

func (s *MultiReaderTestSuite) SetupTest() {
	r1 := streaming.NopCloser(strings.NewReader("abc"))
	r2 := streaming.NopCloser(strings.NewReader("def"))
	r3 := streaming.NopCloser(strings.NewReader("ghi"))

	s.m = &multiReader{
		readers: []io.ReadSeekCloser{r1, r2, r3},
		sizes:   []int64{3, 3, 3},
	}
}

func (s *MultiReaderTestSuite) TestRead() {
	t := s.T()

	buf := make([]byte, 4)
	n, err := s.m.Read(buf)
	require.NoError(t, err)
	require.Equal(t, 4, n)
	assert.EqualValues(t, "abcd", buf)

	n, err = s.m.Read(buf)
	require.NoError(t, err)
	require.Equal(t, 4, n)
	assert.EqualValues(t, "efgh", buf)

	n, err = s.m.Read(buf)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	assert.EqualValues(t, "i", buf[:n])

	n, err = s.m.Read(buf)
	assert.Equal(t, 0, n)
	assert.ErrorIs(t, io.EOF, err)
}

func (s *MultiReaderTestSuite) TestSeek() {
	t := s.T()
	buf := make([]byte, 4)

	// Test Seek io.SeekStart
	off, err := s.m.Seek(1, io.SeekStart)
	require.NoError(t, err)
	assert.EqualValues(t, 1, off)

	n, err := s.m.Read(buf)
	require.NoError(t, err)
	require.Equal(t, 4, n)
	assert.EqualValues(t, "bcde", buf)

	// Test Seek io.SeekCurrent
	off, err = s.m.Seek(1, io.SeekCurrent) // current is 1 + 4 = 5. 5 + 1 = 6.
	require.NoError(t, err)
	assert.EqualValues(t, 6, off)

	n, err = s.m.Read(buf)
	require.NoError(t, err)
	require.Equal(t, 3, n)
	assert.EqualValues(t, "ghi", buf[:n])

	n, err = s.m.Read(buf)
	assert.Equal(t, 0, n)
	assert.ErrorIs(t, io.EOF, err)

	// Test Seek io.SeekEnd
	off, err = s.m.Seek(-2, io.SeekEnd) // total is 9. 9 - 2 = 7.
	require.NoError(t, err)
	assert.EqualValues(t, 7, off)

	n, err = s.m.Read(buf)
	require.NoError(t, err)
	require.Equal(t, 2, n)
	assert.EqualValues(t, "hi", buf[:n])
}

func (s *MultiReaderTestSuite) TestSeek_Edge() {
	t := s.T()
	buf := make([]byte, 4)

	off, err := s.m.Seek(0, io.SeekStart)
	require.NoError(t, err)
	assert.EqualValues(t, 0, off)
	assert.EqualValues(t, 0, s.m.idx)

	off, err = s.m.Seek(3, io.SeekStart)
	require.NoError(t, err)
	assert.EqualValues(t, 3, off)
	assert.EqualValues(t, 1, s.m.idx) // move to next file (idx=1) instead of EOF file idx=0

	off, err = s.m.Seek(0, io.SeekEnd)
	require.NoError(t, err)
	assert.EqualValues(t, 9, off)
	assert.EqualValues(t, 2, s.m.idx)

	_, err = s.m.Read(buf)
	assert.ErrorIs(t, io.EOF, err)

	off, err = s.m.Seek(6, io.SeekCurrent) // 9 (End) + 6 = 15
	require.NoError(t, err)
	assert.EqualValues(t, 15, off)
	assert.EqualValues(t, 2, s.m.idx)

	_, err = s.m.Read(buf)
	assert.ErrorIs(t, io.EOF, err)
}
