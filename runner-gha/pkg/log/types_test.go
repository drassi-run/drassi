package log

import (
	"io"
	"os"
	"path/filepath"
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

type SectionTestSuite struct {
	suite.Suite
	tempFile string
	lines    []string
}

func TestSectionTestSuite(t *testing.T) {
	suite.Run(t, new(SectionTestSuite))
}

func (s *SectionTestSuite) SetupSuite() {
	f, err := os.CreateTemp("", "section-test-*.log")
	s.Require().NoError(err)
	s.tempFile = f.Name()
	s.lines = []string{
		"line1",
		"line2",
		"line3",
		"line4",
		"line5",
	}
	_, err = f.WriteString(strings.Join(s.lines, "\n"))
	s.Require().NoError(err)
	f.Close()
}

func (s *SectionTestSuite) TearDownSuite() {
	os.Remove(s.tempFile)
}

func (s *SectionTestSuite) TestMetadata() {
	sec := &Section{
		startOffset: 10,
		endOffset:   25,
		startLine:   2,
		endLine:     5,
	}

	s.EqualValues(15, sec.Size())
	s.Equal(3, sec.Lines())
}

func (s *SectionTestSuite) TestReader_Full() {
	content := strings.Join(s.lines, "\n")

	// Full file section
	sec := &Section{
		filePath:    s.tempFile,
		startOffset: 0,
		endOffset:   int64(len(content)),
		startLine:   0,
		endLine:     len(s.lines),
	}

	r, err := sec.Reader()
	s.Require().NoError(err)
	defer r.Close()

	buf, err := io.ReadAll(r)
	s.Require().NoError(err)
	s.EqualValues(content, buf)
	s.True(sec.EOF())
}

func (s *SectionTestSuite) TestReader_Partial() {
	// Partial section
	sec := &Section{
		filePath:    s.tempFile,
		startOffset: 6,  // after "line1\n"
		endOffset:   24, // includes "line4\n"
		startLine:   1,
		endLine:     4,
	}

	r, err := sec.Reader()
	s.Require().NoError(err)
	defer r.Close()

	buf, err := io.ReadAll(r)
	s.Require().NoError(err)
	s.EqualValues(strings.Join(s.lines[1:4], "\n")+"\n", buf)
	s.False(sec.EOF())
	s.EqualValues(3, sec.Lines())
}

func (s *SectionTestSuite) TestReader_NonExistentFile() {
	sec := &Section{
		filePath: "non-existent-file.log",
	}

	_, err := sec.Reader()
	s.Error(err)
	s.True(os.IsNotExist(err))
}

type ChunkTestSuite struct {
	suite.Suite
	tmpDir string
	f1     string
	f2     string
}

func TestChunkTestSuite(t *testing.T) {
	suite.Run(t, new(ChunkTestSuite))
}

func (s *ChunkTestSuite) SetupTest() {
	var err error
	s.tmpDir, err = os.MkdirTemp("", "chunk-test")
	s.Require().NoError(err)

	s.f1 = filepath.Join(s.tmpDir, "f1.log")
	err = os.WriteFile(s.f1, []byte("line1\nline2\n"), 0644)
	s.Require().NoError(err)

	s.f2 = filepath.Join(s.tmpDir, "f2.log")
	err = os.WriteFile(s.f2, []byte("line3\nline4\nline5\n"), 0644)
	s.Require().NoError(err)
}

func (s *ChunkTestSuite) TearDownTest() {
	os.RemoveAll(s.tmpDir)
}

func (s *ChunkTestSuite) TestEmpty() {
	assert.True(s.T(), Chunk{}.Empty())
	assert.False(s.T(), Chunk{&Section{}}.Empty())
}

func (s *ChunkTestSuite) TestSize() {
	c := Chunk{
		{startOffset: 0, endOffset: 12},
		{startOffset: 0, endOffset: 18},
	}
	assert.EqualValues(s.T(), 30, c.Size())
}

func (s *ChunkTestSuite) TestLines() {
	c := Chunk{
		{startLine: 0, endLine: 2},
		{startLine: 0, endLine: 3},
	}
	assert.EqualValues(s.T(), 5, c.Lines())
}

func (s *ChunkTestSuite) TestReader_Empty() {
	r, err := Chunk{}.Reader()
	require.NoError(s.T(), err)
	defer r.Close()

	content, err := io.ReadAll(r)
	require.NoError(s.T(), err)
	assert.Empty(s.T(), content)
}

func (s *ChunkTestSuite) TestReader_Single() {
	s1 := &Section{filePath: s.f1, startOffset: 0, endOffset: 12, startLine: 0, endLine: 2}
	r, err := Chunk{s1}.Reader()
	s.Require().NoError(err)
	defer r.Close()

	content, err := io.ReadAll(r)
	s.Require().NoError(err)
	s.EqualValues("line1\nline2\n", content)
}

func (s *ChunkTestSuite) TestReader_Multiple() {
	s1 := &Section{filePath: s.f1, startOffset: 0, endOffset: 12, startLine: 0, endLine: 2}
	s2 := &Section{filePath: s.f2, startOffset: 0, endOffset: 18, startLine: 0, endLine: 3}
	r, err := Chunk{s1, s2}.Reader()
	s.Require().NoError(err)
	defer r.Close()

	content, err := io.ReadAll(r)
	s.Require().NoError(err)
	s.EqualValues("line1\nline2\nline3\nline4\nline5\n", content)
}
