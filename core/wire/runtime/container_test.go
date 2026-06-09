/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_runtime

import (
	"path"
	"testing"

	"drassi.run/core/pkg/container/types"
	"github.com/stretchr/testify/assert"
)

func TestChMount(t *testing.T) {
	t.Run("tmp", func(t *testing.T) {
		m := &types.Mount{
			Type:   "tmp",
			Target: "/tmp",
		}
		_, err := chMount(m, "/tmp/foobar", "")
		assert.ErrorContains(t, err, "unsupported mount type")
	})

	t.Run("bind", func(t *testing.T) {
		source := "/path/to/host"
		mountPath := "/path/to/container"
		m := &types.Mount{
			Type:   "bind",
			Source: source,
			Target: "/path/to/sandbox",
		}
		mount, err := chMount(m, mountPath, "")
		assert.NoError(t, err)
		assert.Equal(t, "bind", mount.Type)
		assert.Equal(t, source, mount.Source)
		assert.Equal(t, mountPath, mount.Target)

		mount, err = chMount(m, mountPath, "subdir")
		assert.NoError(t, err)
		assert.Equal(t, "bind", mount.Type)
		assert.Equal(t, path.Join(source, "subdir"), mount.Source)
		assert.Equal(t, mountPath, mount.Target)
	})

	t.Run("volume", func(t *testing.T) {
		mountPath := "/path/to/container"
		m := &types.Mount{
			Type:   "volume",
			Source: "vol1",
			Target: "/path/to/sandbox",
		}
		mount, err := chMount(m, mountPath, "")
		assert.NoError(t, err)
		assert.Equal(t, "volume", mount.Type)
		assert.Equal(t, "vol1", mount.Source)
		assert.Equal(t, mountPath, mount.Target)

		mount, err = chMount(m, mountPath, "subdir")
		assert.NoError(t, err)
		assert.Equal(t, "volume", mount.Type)
		assert.Equal(t, "vol1", mount.Source)
		assert.Equal(t, mountPath, mount.Target)
		assert.Equal(t, "subdir", mount.VolumeOptions.SubPath)

		m.VolumeOptions = &types.VolumeOptions{
			SubPath: "foo/bar",
			Driver:  "local",
		}
		mount, err = chMount(m, mountPath, "subdir")
		assert.NoError(t, err)
		assert.Equal(t, "foo/bar/subdir", mount.VolumeOptions.SubPath)
		assert.Equal(t, "local", mount.VolumeOptions.Driver)
	})
}

var (
	m1 = &types.Mount{
		Type:   "volume",
		Source: "vol1",
		Target: "/path/to/foo",
	}
	m2 = &types.Mount{
		Type:   "bind",
		Source: "/does/not/matter",
		Target: "/path/to/bar/",
	}
	m3 = &types.Mount{
		Type:   "tmp",
		Target: "/path/to/",
	}
	m4 = &types.Mount{
		Type:   "volume",
		Source: "vol2",
		Target: "/a/new/path",
		VolumeOptions: &types.VolumeOptions{
			SubPath: "dir1",
		},
	}
	mounts = []*types.Mount{m1, m2, m3, m4}
)

func TestMountOf(t *testing.T) {
	t.Run("exact match", func(t *testing.T) {
		for _, mount := range mounts {
			m, p := mountOf(mount.Target, mounts)
			assert.Equal(t, mount, m)
			assert.Equal(t, "", p)
		}
	})

	t.Run("no match", func(t *testing.T) {
		m, _ := mountOf("/not/exist", mounts)
		assert.Nil(t, m)
	})

	t.Run("subdir", func(t *testing.T) {
		type testcase struct {
			path   string
			mount  *types.Mount
			subdir string
		}
		tests := []testcase{
			{"/path/to/foo/hello", m1, "hello"},
			{"/path/to/abcxyz/", m3, "abcxyz/"},
			{"/path/to/f", m3, "f"},                         // f is prefix of foo
			{"/path/to/foooooooooooo", m3, "foooooooooooo"}, // foo is prefix of foooooooooooo
			{"/a/new/path/subdir", m4, "subdir"},
		}
		for _, tc := range tests {
			m, p := mountOf(tc.path, mounts)
			assert.Equal(t, tc.mount, m)
			assert.Equal(t, tc.subdir, p)
		}
	})
}
