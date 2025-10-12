package log

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

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
	s.EqualValues(3, sec.Lines())
}

func (s *SectionTestSuite) TestReader_Full() {
	content := strings.Join(s.lines, "\n")

	// Full file section
	sec := &Section{
		filePath:    s.tempFile,
		startOffset: 0,
		endOffset:   int64(len(content)),
		startLine:   0,
		endLine:     int64(len(s.lines)),
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
