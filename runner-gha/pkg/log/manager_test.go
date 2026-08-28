/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package log

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

var tsSize = len(time.Now().UTC().Format(RFC3339Tick)) + 1

func tsTrim(b []byte) string {
	lines := make([]string, 0)
	for _, l := range strings.Split(string(b), "\n") {
		if l == "" {
			lines = append(lines, "")
			continue
		}
		lines = append(lines, l[tsSize:])
	}
	return strings.Join(lines, "\n")
}

func (s *ManagerTestSuite) SetupTest() {
	var err error
	tempDir, err := os.MkdirTemp("", "log-manager-*")
	s.Require().NoError(err)

	s.m, err = NewManager(tempDir, int64(tsSize*2+25))
	s.Require().NoError(err)
}

func (s *ManagerTestSuite) TearDownTest() {
	err := s.m.Close()
	s.Require().NoError(err)
}

func (s *ManagerTestSuite) TestStartStop() {
	t := s.T()

	// Test Start multiple distinct sessions concurrently
	uid1 := "session-1"
	uid2 := "session-2"
	err := s.m.Start(uid1)
	require.NoError(t, err)
	err = s.m.Start(uid2)
	require.NoError(t, err)

	// Test Start again with existing uid (should error)
	err = s.m.Start(uid1)
	assert.ErrorContains(t, err, "already started")

	// Test Stop uid1
	err = s.m.Stop(uid1)
	require.NoError(t, err)

	// Test Stop uid1 again (should be idempotent)
	err = s.m.Stop(uid1)
	require.NoError(t, err)

	// Verify uid2 is still active and can be stopped
	err = s.m.Stop(uid2)
	require.NoError(t, err)
}

func (s *ManagerTestSuite) TestHandle_ConcurrentSteps() {
	t := s.T()
	s.m.maxSize = 1024
	uid1 := "step-1"
	uid2 := "step-2"

	require.NoError(t, s.m.Start(uid1))
	require.NoError(t, s.m.Start(uid2))

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			err := s.m.Handle(uid1, fmt.Sprintf("line-1-%d", i))
			assert.NoError(t, err)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			err := s.m.Handle(uid2, fmt.Sprintf("line-2-%d", i))
			assert.NoError(t, err)
		}
	}()

	wg.Wait()

	require.NoError(t, s.m.Stop(uid1))
	require.NoError(t, s.m.Stop(uid2))

	logFile1 := filepath.Join(s.m.dir, uid1+".0.log")
	content1, err := os.ReadFile(logFile1)
	require.NoError(t, err)
	expected1 := "line-1-0\nline-1-1\nline-1-2\nline-1-3\nline-1-4\n"
	assert.Equal(t, expected1, tsTrim(content1))

	logFile2 := filepath.Join(s.m.dir, uid2+".0.log")
	content2, err := os.ReadFile(logFile2)
	require.NoError(t, err)
	expected2 := "line-2-0\nline-2-1\nline-2-2\nline-2-3\nline-2-4\n"
	assert.Equal(t, expected2, tsTrim(content2))
}

func (s *ManagerTestSuite) TestHandle_UnstartedSession() {
	t := s.T()
	// Unstarted session logs should be dropped without error
	err := s.m.Handle("unstarted-session", "dropped line")
	require.NoError(t, err)

	// Verify no file was created
	files, err := os.ReadDir(s.m.dir)
	require.NoError(t, err)
	assert.Empty(t, files)
}

func (s *ManagerTestSuite) TestHandle_Rotation() {
	t := s.T()
	uid := "rotate-session"
	err := s.m.Start(uid)
	require.NoError(t, err)

	// Write first 2 lines - should NOT trigger rotation yet
	line := strings.Repeat("x", 10)
	for range 2 {
		err = s.m.Handle(uid, line)
		require.NoError(t, err)
	}

	// 3rd line -> trigger rotate
	err = s.m.Handle(uid, line)
	require.NoError(t, err)

	// 4th line -> written in next file (.1.log)
	err = s.m.Handle(uid, line)
	require.NoError(t, err)

	require.NoError(t, s.m.Stop(uid))

	logFile0 := filepath.Join(s.m.dir, uid+".0.log")
	content0, err := os.ReadFile(logFile0)
	require.NoError(t, err)
	assert.Equal(t, strings.Repeat(line+"\n", 3), tsTrim(content0))

	logFile1 := filepath.Join(s.m.dir, uid+".1.log")
	content1, err := os.ReadFile(logFile1)
	require.NoError(t, err)
	assert.Equal(t, line+"\n", tsTrim(content1))
}

func (s *ManagerTestSuite) TestSubscribe() {
	t := s.T()
	sub := s.m.Subscribe()

	uid := "sub-step"
	err := s.m.Start(uid)
	require.NoError(t, err)

	select {
	case e := <-sub:
		assert.Equal(t, OnRecordStart, e.Kind)
		assert.Equal(t, uid, e.Uid)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for OnRecordStart")
	}

	line := strings.Repeat("x", 15)
	err = s.m.Handle(uid, line)
	require.NoError(t, err)

	size := tsSize + len(line) + 1
	select {
	case e := <-sub:
		assert.Equal(t, OnRecordLog, e.Kind)
		assert.Equal(t, uid, e.Uid)
		assert.Equal(t, filepath.Join(s.m.dir, uid+".0.log"), e.File)
		assert.False(t, e.Complete)
		assert.Equal(t, 1, e.Line)
		assert.EqualValues(t, size, e.Offset)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for OnRecordLog")
	}

	err = s.m.Stop(uid)
	require.NoError(t, err)

	select {
	case e := <-sub:
		assert.Equal(t, OnRecordStop, e.Kind)
		assert.Equal(t, uid, e.Uid)
		assert.True(t, e.Complete)
		assert.Equal(t, 1, e.Line)
		assert.EqualValues(t, size, e.Offset)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for OnRecordStop")
	}
}

func (s *ManagerTestSuite) TestClose() {
	t := s.T()
	sub := s.m.Subscribe()

	err := s.m.Start("active-session")
	require.NoError(t, err)

	err = s.m.Close()
	require.NoError(t, err)

	// Verify channel is closed
	for range sub {
	}
	_, ok := <-sub
	assert.False(t, ok)
}
