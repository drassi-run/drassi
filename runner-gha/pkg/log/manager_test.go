/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package log

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestNewManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "log-manager-*")
	defer os.RemoveAll(tempDir)
	require.NoError(t, err)

	maxSize := int64(1024)
	basePath := filepath.Join(tempDir, "job-1")
	m, err := NewManager(basePath, maxSize)
	require.NoError(t, err)
	assert.NotNil(t, m)
	assert.Equal(t, basePath, m.dir)
	assert.Equal(t, maxSize, m.maxSize)

	// Verify directory was created
	info, err := os.Stat(basePath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestManagerSuite(t *testing.T) {
	suite.Run(t, new(ManagerTestSuite))
}

type ManagerTestSuite struct {
	suite.Suite
	m *Manager
}

func (s *ManagerTestSuite) SetupTest() {
	var err error
	tempDir, err := os.MkdirTemp("", "log-manager-*")
	s.Require().NoError(err)

	s.m, err = NewManager(tempDir, 25)
	s.Require().NoError(err)
}

func (s *ManagerTestSuite) TearDownTest() {
	err := s.m.Stop()
	s.Require().NoError(err)

	err = s.m.Close()
	s.Require().NoError(err)
}

func (s *ManagerTestSuite) TestStartStop() {
	t := s.T()

	// Test Start
	uid := "test-session"
	err := s.m.Start(uid)
	require.NoError(t, err)
	assert.Equal(t, uid, s.m.currUid)

	// Test Start again (should error)
	err = s.m.Start("another-uid")
	assert.ErrorContains(t, err, "session already started")

	// Test Stop
	err = s.m.Stop()
	require.NoError(t, err)
	assert.Empty(t, s.m.currUid)

	// Test Stop again (should be idempotent)
	err = s.m.Stop()
	require.NoError(t, err)
}

func (s *ManagerTestSuite) TestHandle_Write() {
	t := s.T()
	uid := "session-1"
	err := s.m.Start(uid)
	require.NoError(t, err)

	// write log "hello"
	line := "hello"
	err = s.m.Handle(t.Context(), line)
	require.NoError(t, err)
	assert.EqualValues(t, 1, s.m.currLines)
	size := len(line) + 1
	assert.EqualValues(t, size, s.m.currSize)

	logFile := filepath.Join(s.m.dir, uid+".0.log")
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)
	assert.Equal(t, "hello\n", string(content))

	// write log "world"
	line = "world"
	err = s.m.Handle(t.Context(), line)
	require.NoError(t, err)
	assert.EqualValues(t, 2, s.m.currLines)
	size += len(line) + 1
	assert.EqualValues(t, size, s.m.currSize)

	content, err = os.ReadFile(logFile)
	require.NoError(t, err)
	assert.Equal(t, "hello\nworld\n", string(content))
}

func (s *ManagerTestSuite) TestHandle_Rotation() {
	t := s.T()
	uid := "rotate-session"
	err := s.m.Start(uid)
	require.NoError(t, err)

	// Write first 2 lines - should NOT trigger rotation yet
	line := strings.Repeat("x", 10) // 10 bytes + 1 newline = 11 bytes
	for range 2 {                   // 2 lines = 22 bytes
		err = s.m.Handle(t.Context(), line)
		require.NoError(t, err)
		assert.Equal(t, 0, s.m.idx)
		assert.NotNil(t, s.m.f)
	}

	// 3rd line -> trigger rotate
	err = s.m.Handle(t.Context(), line)
	require.NoError(t, err)
	assert.Equal(t, 1, s.m.idx)
	assert.Nil(t, s.m.f)
	assert.EqualValues(t, 0, s.m.currLines)
	assert.EqualValues(t, 0, s.m.currSize)

	// 4th line -> in next file
	err = s.m.Handle(t.Context(), line)
	require.NoError(t, err)
	assert.Equal(t, 1, s.m.idx)
	assert.NotNil(t, s.m.f)
	assert.EqualValues(t, 1, s.m.currLines)
	assert.EqualValues(t, len(line)+1, s.m.currSize)

	logFile := filepath.Join(s.m.dir, uid+".1.log")
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)
	assert.Equal(t, line+"\n", string(content))
}

func (s *ManagerTestSuite) TestSubscribe() {
	t := s.T()
	sub := s.m.Subscribe()

	share := func(uid string) {
		// OnRecordStart
		err := s.m.Start(uid)
		require.NoError(t, err)

		select {
		case e := <-sub:
			assert.Equal(t, OnRecordStart, e.Kind)
			assert.Equal(t, uid, e.Uid)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for OnRecordStart")
		}

		// 1st OnRecordLog
		line := strings.Repeat("x", 15)
		err = s.m.Handle(t.Context(), line)
		require.NoError(t, err)

		size := len(line) + 1
		select {
		case e := <-sub:
			assert.Equal(t, OnRecordLog, e.Kind)
			assert.Equal(t, uid, e.Uid)
			assert.Equal(t, filepath.Join(s.m.dir, uid+".0.log"), e.File)
			assert.False(t, e.Complete)
			assert.EqualValues(t, 1, e.Line)
			assert.EqualValues(t, size, e.Offset)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for OnRecordLog")
		}

		// 2nd OnRecordLog - file size reach limit => rotate
		err = s.m.Handle(t.Context(), line)
		require.NoError(t, err)

		size += len(line) + 1
		select {
		case e := <-sub:
			assert.Equal(t, OnRecordLog, e.Kind)
			assert.Equal(t, uid, e.Uid)
			assert.Equal(t, filepath.Join(s.m.dir, uid+".0.log"), e.File)
			assert.True(t, e.Complete)
			assert.EqualValues(t, 2, e.Line)
			assert.EqualValues(t, size, e.Offset)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for OnRecordLog")
		}
	}

	step1 := func() {
		uid := "step1"
		share(uid)

		// OnRecordStop
		err := s.m.Stop()
		require.NoError(t, err)

		select {
		case e := <-sub:
			assert.Equal(t, OnRecordStop, e.Kind)
			assert.Equal(t, uid, e.Uid)
			assert.Nil(t, e.Update)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for OnRecordStop")
		}
	}

	step2 := func() {
		uid := "step2"
		share(uid)

		// 3rd OnRecordLog - write in 2nd file
		line := strings.Repeat("z", 20)
		err := s.m.Handle(t.Context(), line)
		require.NoError(t, err)

		size := len(line) + 1
		file := filepath.Join(s.m.dir, uid+".1.log")
		select {
		case e := <-sub:
			assert.Equal(t, OnRecordLog, e.Kind)
			assert.Equal(t, uid, e.Uid)
			assert.Equal(t, file, e.File)
			assert.False(t, e.Complete)
			assert.EqualValues(t, 1, e.Line)
			assert.EqualValues(t, size, e.Offset)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for OnRecordLog")
		}

		// OnRecordStop
		err = s.m.Stop()
		require.NoError(t, err)

		select {
		case e := <-sub:
			assert.Equal(t, OnRecordStop, e.Kind)
			assert.Equal(t, uid, e.Uid)
			assert.Equal(t, file, e.File)
			assert.True(t, e.Complete)
			assert.EqualValues(t, 1, e.Line)
			assert.EqualValues(t, size, e.Offset)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for OnRecordStop")
		}
	}

	step1()
	step2()
}

func (s *ManagerTestSuite) TestClose() {
	t := s.T()
	sub := s.m.Subscribe()

	// Cannot close active session
	err := s.m.Start("active")
	require.NoError(t, err)

	err = s.m.Close()
	assert.ErrorContains(t, err, "must stop before close")

	err = s.m.Stop()
	require.NoError(t, err)
	err = s.m.Close()
	require.NoError(t, err)

	// Verify channel is closed
	for range sub {
	}
	_, ok := <-sub
	assert.False(t, ok)
}
