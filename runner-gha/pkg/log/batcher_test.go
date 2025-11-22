/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package log

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/suite"
)

func TestBatcherTestSuite(t *testing.T) {
	suite.Run(t, new(BatcherTestSuite))
}

type BatcherTestSuite struct {
	suite.Suite
}

func (s *BatcherTestSuite) TestThresholdTrigger() {
	synctest.Test(s.T(), func(t *testing.T) {
		// Interval is very long, so only threshold can trigger
		b := NewBatcher(10, time.Hour)
		defer b.Close()

		for i := 0; i < 9; i++ {
			b.Update(&Update{File: "test.log", Offset: int64((i + 1) * 10), Line: i + 1})
		}

		// Should not emit yet
		select {
		case <-b.Channel():
			t.Fatal("Did not expect a batch yet")
		default:
		}

		// 10th update reaches threshold
		b.Update(&Update{File: "test.log", Offset: 100, Line: 10})

		// Wait for goroutines to process and emit
		synctest.Wait()

		select {
		case batch := <-b.Channel():
			s.Require().NotNil(batch)
			s.EqualValues(10, batch.Lines())
		default:
			t.Fatal("Expected a batch to be emitted")
		}
	})
}

func (s *BatcherTestSuite) TestIntervalTrigger() {
	synctest.Test(s.T(), func(t *testing.T) {
		interval := time.Second
		b := NewBatcher(100, interval)
		defer b.Close()

		b.Update(&Update{File: "test.log", Offset: 10, Line: 1, Complete: true})

		// Should not emit yet
		select {
		case <-b.Channel():
			t.Fatal("Did not expect a batch yet")
		default:
		}

		// Advance time beyond interval
		time.Sleep(interval + 100*time.Millisecond)
		synctest.Wait()

		select {
		case batch := <-b.Channel():
			s.Require().NotNil(batch)
			s.EqualValues(1, batch.Lines())
		default:
			t.Fatal("Expected a batch to be emitted after interval")
		}
	})
}

func (s *BatcherTestSuite) TestFlushOnClose() {
	synctest.Test(s.T(), func(t *testing.T) {
		b := NewBatcher(100, time.Hour)
		b.Update(&Update{File: "test.log", Offset: 50, Line: 5})

		err := b.Close()
		s.Require().NoError(err)
		synctest.Wait()

		select {
		case batch := <-b.Channel():
			s.Require().NotNil(batch)
			s.EqualValues(5, batch.Lines())
		default:
			t.Fatal("Expected remaining batch to be flushed on Close")
		}

		_, ok := <-b.Channel()
		s.False(ok, "Channel should be closed after Close()")
	})
}

func (s *BatcherTestSuite) TestMultipleSectionsInBatch() {
	synctest.Test(s.T(), func(t *testing.T) {
		b := NewBatcher(10, time.Hour)
		defer b.Close()

		// Section 1: file1, 4 lines, completed
		b.Update(&Update{File: "file1.log", Offset: 40, Line: 4, Complete: true})
		// Section 2: file2, 6 lines
		b.Update(&Update{File: "file2.log", Offset: 60, Line: 6})

		synctest.Wait()

		select {
		case batch := <-b.Channel():
			s.Require().NotNil(batch)
			s.EqualValues(10, batch.Lines())

			bk := batch.(sections)
			s.Len(bk, 2)
			s.Equal("file1.log", bk[0].filePath)
			s.Equal("file2.log", bk[1].filePath)
		default:
			t.Fatal("Expected batch with multiple sections")
		}
	})
}

func (s *BatcherTestSuite) TestThresholdWithPreExistingTotal() {
	synctest.Test(s.T(), func(t *testing.T) {
		b := NewBatcher(10, time.Hour)
		defer b.Close()

		// First section, 5 lines
		b.Update(&Update{File: "file1.log", Offset: 50, Line: 5, Complete: true})
		// Second section, starts with 5 lines, then update with 5 more
		b.Update(&Update{File: "file2.log", Offset: 50, Line: 5})
		b.Update(&Update{File: "file2.log", Offset: 100, Line: 10}) // Reaches total 15, should split

		synctest.Wait()

		select {
		case batch := <-b.Channel():
			s.Require().NotNil(batch)
			s.EqualValues(15, batch.Lines())
		default:
			t.Fatal("Expected batch when threshold reached via update")
		}
	})
}
