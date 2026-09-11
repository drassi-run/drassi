/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package parser_test

import (
	"testing"

	"drassi.run/core/pkg/container/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseVolume(t *testing.T) {
	t.Run("bind mount ro", func(t *testing.T) {
		m, err := parser.ParseVolume("/source:/target:ro")
		require.NoError(t, err)
		assert.Equal(t, "bind", m.Type)
		assert.Equal(t, "/source", m.Source)
		assert.Equal(t, "/target", m.Target)
		assert.True(t, m.ReadOnly)
	})

	t.Run("volume rw", func(t *testing.T) {
		m, err := parser.ParseVolume("my-volume:/target:rw")
		require.NoError(t, err)
		assert.Equal(t, "volume", m.Type)
		assert.Equal(t, "my-volume", m.Source)
		assert.Equal(t, "/target", m.Target)
		assert.False(t, m.ReadOnly)
	})

	t.Run("invalid empty", func(t *testing.T) {
		_, err := parser.ParseVolume("")
		assert.Error(t, err)
	})
}

func TestParseTmpfs(t *testing.T) {
	t.Run("simple tmpfs", func(t *testing.T) {
		m, err := parser.ParseTmpfs("/tmp")
		require.NoError(t, err)
		assert.Equal(t, "tmpfs", m.Type)
		assert.Equal(t, "/tmp", m.Target)
		assert.Nil(t, m.TmpfsOptions)
	})

	t.Run("tmpfs with options", func(t *testing.T) {
		m, err := parser.ParseTmpfs("/tmp:size=64m,ro,mode=1777")
		require.NoError(t, err)
		assert.Equal(t, "tmpfs", m.Type)
		assert.Equal(t, "/tmp", m.Target)
		assert.True(t, m.ReadOnly)
		require.NotNil(t, m.TmpfsOptions)
		assert.Equal(t, int64(64*1024*1024), m.TmpfsOptions.Size)
		assert.Equal(t, uint32(01777), uint32(m.TmpfsOptions.Mode))
	})

	t.Run("invalid relative target", func(t *testing.T) {
		_, err := parser.ParseTmpfs("relative/path")
		assert.Error(t, err)
	})
}
