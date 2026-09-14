/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseVolume(t *testing.T) {
	t.Run("bind mount ro", func(t *testing.T) {
		m, err := ParseVolume("/source:/target:ro")
		require.NoError(t, err)
		assert.Equal(t, "bind", m.Type)
		assert.Equal(t, "/source", m.Source)
		assert.Equal(t, "/target", m.Target)
		assert.True(t, m.ReadOnly)
	})

	t.Run("volume rw", func(t *testing.T) {
		m, err := ParseVolume("my-volume:/target:rw")
		require.NoError(t, err)
		assert.Equal(t, "volume", m.Type)
		assert.Equal(t, "my-volume", m.Source)
		assert.Equal(t, "/target", m.Target)
		assert.False(t, m.ReadOnly)
	})

	t.Run("volume nocopy", func(t *testing.T) {
		m, err := ParseVolume("my-volume:/target:nocopy")
		require.NoError(t, err)
		assert.Equal(t, "volume", m.Type)
		assert.Equal(t, "my-volume", m.Source)
		assert.Equal(t, "/target", m.Target)
		require.NotNil(t, m.VolumeOptions)
		assert.True(t, m.VolumeOptions.NoCopy)
	})

	t.Run("bind mount with propagation and consistency", func(t *testing.T) {
		m, err := ParseVolume("/source:/target:rslave,cached")
		require.NoError(t, err)
		assert.Equal(t, "bind", m.Type)
		assert.Equal(t, "/source", m.Source)
		assert.Equal(t, "/target", m.Target)
		require.NotNil(t, m.BindOptions)
		assert.Equal(t, "rslave", m.BindOptions.Propagation)
		assert.Equal(t, "cached", m.BindOptions.Consistency)
	})

	t.Run("anonymous volume unix", func(t *testing.T) {
		m, err := ParseVolume("/data")
		require.NoError(t, err)
		assert.Equal(t, "volume", m.Type)
		assert.Empty(t, m.Source)
		assert.Equal(t, "/data", m.Target)
	})

	t.Run("anonymous volume windows", func(t *testing.T) {
		m, err := ParseVolume(`C:\data`)
		require.NoError(t, err)
		assert.Equal(t, "volume", m.Type)
		assert.Empty(t, m.Source)
		assert.Equal(t, `C:\data`, m.Target)
	})

	t.Run("windows bind mount with drive letters", func(t *testing.T) {
		m, err := ParseVolume(`C:\source\foo:D:\target:ro,rprivate`)
		require.NoError(t, err)
		assert.Equal(t, "bind", m.Type)
		assert.Equal(t, `C:\source\foo`, m.Source)
		assert.Equal(t, `D:\target`, m.Target)
		assert.True(t, m.ReadOnly)
		require.NotNil(t, m.BindOptions)
		assert.Equal(t, "rprivate", m.BindOptions.Propagation)
	})

	t.Run("windows named pipe", func(t *testing.T) {
		m, err := ParseVolume(`\\.\pipe\docker_engine:\\.\pipe\inside`)
		require.NoError(t, err)
		assert.Equal(t, "bind", m.Type)
		assert.Equal(t, `\\.\pipe\docker_engine`, m.Source)
		assert.Equal(t, `\\.\pipe\inside`, m.Target)
	})

	t.Run("invalid empty", func(t *testing.T) {
		_, err := ParseVolume("")
		assert.Error(t, err)
	})

	t.Run("invalid too many colons", func(t *testing.T) {
		_, err := ParseVolume("/foo:/bar:ro:extra")
		assert.ErrorContains(t, err, "too many colons")
	})

	t.Run("invalid empty section", func(t *testing.T) {
		for _, spec := range []string{":foo", "/foo::ro", "/foo:"} {
			_, err := ParseVolume(spec)
			assert.ErrorContains(t, err, "empty section between colons")
		}
	})
}

func TestParseTmpfs(t *testing.T) {
	t.Run("simple tmpfs", func(t *testing.T) {
		m, err := ParseTmpfs("/tmp")
		require.NoError(t, err)
		assert.Equal(t, "tmpfs", m.Type)
		assert.Equal(t, "/tmp", m.Target)
		assert.Nil(t, m.TmpfsOptions)
	})

	t.Run("tmpfs with options", func(t *testing.T) {
		m, err := ParseTmpfs("/tmp:size=64m,ro,mode=1777")
		require.NoError(t, err)
		assert.Equal(t, "tmpfs", m.Type)
		assert.Equal(t, "/tmp", m.Target)
		assert.True(t, m.ReadOnly)
		require.NotNil(t, m.TmpfsOptions)
		assert.Equal(t, int64(64*1024*1024), m.TmpfsOptions.Size)
		assert.Equal(t, uint32(01777), uint32(m.TmpfsOptions.Mode))
	})

	t.Run("invalid relative target", func(t *testing.T) {
		_, err := ParseTmpfs("relative/path")
		assert.Error(t, err)
	})
}
