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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
)

func TestSession_LazyFileCreation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "session-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	s := newSession(tempDir, "step-1", attribute.String("key", "val"))
	assert.Equal(t, filepath.Join(tempDir, "step-1.0.log"), s.filePath())

	// No file before write
	_, err = os.Stat(s.filePath())
	assert.True(t, os.IsNotExist(err))

	// Write creates the file
	u, err := s.Write("hello world", 1024)
	require.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, s.filePath(), u.File)
	assert.False(t, u.Complete)
	assert.Equal(t, 1, u.Line)

	_, err = os.Stat(s.filePath())
	assert.NoError(t, err)
}

func TestSession_RotationAndStop(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "session-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	s := newSession(tempDir, "step-rot")
	line := strings.Repeat("a", 10)
	// line size is tsSize + 10 + 1 (~47 bytes)
	maxSize := int64(100)

	// First write: ~47 bytes, no rotation
	u1, err := s.Write(line, maxSize)
	require.NoError(t, err)
	assert.False(t, u1.Complete)
	assert.Equal(t, filepath.Join(tempDir, "step-rot.0.log"), u1.File)

	// Second write: ~94 bytes, no rotation
	u2, err := s.Write(line, maxSize)
	require.NoError(t, err)
	assert.False(t, u2.Complete)
	assert.Equal(t, filepath.Join(tempDir, "step-rot.0.log"), u2.File)

	// Third write: ~141 bytes >= 100, rotates after setting complete
	u3, err := s.Write(line, maxSize)
	require.NoError(t, err)
	assert.True(t, u3.Complete)
	assert.Equal(t, filepath.Join(tempDir, "step-rot.0.log"), u3.File)

	// Next write should go to step-rot.1.log
	u4, err := s.Write(line, 1000)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tempDir, "step-rot.1.log"), u4.File)
	assert.False(t, u4.Complete)

	// Stop finishes step-rot.1.log
	stopUpdate, err := s.Stop()
	require.NoError(t, err)
	require.NotNil(t, stopUpdate)
	assert.True(t, stopUpdate.Complete)
	assert.Equal(t, filepath.Join(tempDir, "step-rot.1.log"), stopUpdate.File)

	// Stopping again when no file is open returns nil update and no error
	stopUpdate2, err := s.Stop()
	require.NoError(t, err)
	assert.Nil(t, stopUpdate2)
}
