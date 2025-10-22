package log

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ChunkerTestSuite struct {
	suite.Suite
	cr *Chunker
}

func TestChunkerTestSuite(t *testing.T) {
	suite.Run(t, new(ChunkerTestSuite))
}

func (s *ChunkerTestSuite) SetupTest() {
	s.cr = NewChunker(100)
}

func (s *ChunkerTestSuite) TestBasicChunking() {
	s.cr.Update(&Update{File: "test.log", Offset: 150, Line: 10})

	select {
	case c := <-s.cr.Channel():
		s.Len(c, 1)
		s.EqualValues(150, c.Size())
		s.Equal("test.log", c[0].filePath)
		s.EqualValues(0, c[0].startOffset)
		s.EqualValues(150, c[0].endOffset)
	default:
		s.FailNow("Expected a chunk to be emitted")
	}
}

func (s *ChunkerTestSuite) TestMultiUpdateChunking() {
	s.cr.Update(&Update{File: "test.log", Offset: 40, Line: 5})
	s.cr.Update(&Update{File: "test.log", Offset: 80, Line: 10})

	select {
	case <-s.cr.Channel():
		s.FailNow("Did not expect a chunk yet")
	default:
	}

	s.cr.Update(&Update{File: "test.log", Offset: 120, Line: 15})

	select {
	case c := <-s.cr.Channel():
		s.Len(c, 3)
		s.EqualValues(120, c.Size())
		s.Equal(15, c.Lines())
	default:
		s.FailNow("Expected a chunk to be emitted")
	}
}

func (s *ChunkerTestSuite) TestCompletionTrigger() {
	s.cr.Update(&Update{File: "test.log", Offset: 50, Line: 5, Complete: true})

	select {
	case <-s.cr.Channel():
		s.FailNow("Did not expect a chunk yet (below limit)")
	default:
	}

	s.cr.Update(&Update{File: "test2.log", Offset: 60, Line: 6, Complete: true})

	select {
	case c := <-s.cr.Channel():
		s.Len(c, 2)
		s.EqualValues(110, c.Size())
		s.Equal(11, c.Lines())
	default:
		s.FailNow("Expected a chunk to be emitted (total size 110 >= 100)")
	}
}

func (s *ChunkerTestSuite) TestFlushOnClose() {
	s.cr.Update(&Update{File: "test.log", Offset: 50, Line: 5})

	err := s.cr.Close()
	s.NoError(err)

	select {
	case c := <-s.cr.Channel():
		s.Len(c, 1)
		s.EqualValues(50, c.Size())
	default:
		s.FailNow("Expected a chunk to be flushed on close")
	}

	_, ok := <-s.cr.Channel()
	s.False(ok, "Channel should be closed")
}

func (s *ChunkerTestSuite) TestMultipleFiles() {
	s.cr.Update(&Update{File: "file1.log", Offset: 60, Line: 5})
	s.cr.Update(&Update{File: "file2.log", Offset: 60, Line: 10}) // Note: Offset in Update is absolute for THAT file

	select {
	case c := <-s.cr.Channel():
		s.Len(c, 2)
		s.EqualValues(120, c.Size())
		s.Equal("file1.log", c[0].filePath)
		s.Equal("file2.log", c[1].filePath)
	default:
		s.FailNow("Expected a chunk to be emitted")
	}
}
