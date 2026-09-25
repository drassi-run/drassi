/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package docker

import (
	"context"
	"io"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFSAdapter_WalkAndOpen(t *testing.T) {
	mockFS := fstest.MapFS{
		"Dockerfile":       &fstest.MapFile{Data: []byte("FROM alpine\n")},
		"src/main.go":      &fstest.MapFile{Data: []byte("package main\n")},
		"src/util/util.go": &fstest.MapFile{Data: []byte("package util\n")},
	}

	adapter := newFSAdapter(mockFS)
	require.NotNil(t, adapter)

	// Test Open
	rc, err := adapter.Open("Dockerfile")
	require.NoError(t, err)
	data, err := io.ReadAll(rc)
	require.NoError(t, err)
	_ = rc.Close()
	assert.Equal(t, "FROM alpine\n", string(data))

	// Test Open with leading slash
	rc, err = adapter.Open("/src/main.go")
	require.NoError(t, err)
	data, err = io.ReadAll(rc)
	require.NoError(t, err)
	_ = rc.Close()
	assert.Equal(t, "package main\n", string(data))

	// Test Walk
	var walked []string
	err = adapter.Walk(context.Background(), "", func(path string, d fs.DirEntry, err error) error {
		require.NoError(t, err)
		walked = append(walked, path)
		return nil
	})
	require.NoError(t, err)
	assert.Contains(t, walked, "Dockerfile")
	assert.Contains(t, walked, "src/main.go")
	assert.Contains(t, walked, "src/util/util.go")
}
