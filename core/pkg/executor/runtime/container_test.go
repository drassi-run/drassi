/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package runtime

import (
	"context"
	"testing"

	mock_container "drassi.run/core/mock/container"
	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/types"
	. "drassi.run/core/util/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewContainerRuntime(t *testing.T) {
	t.Run("mount_duplicated", testNewContainerRuntime_MountDuplicated)
	t.Run("mount_sorted", testNewContainerRuntime_MountSorted)
}

//goland:noinspection GoSnakeCaseUsage
func testNewContainerRuntime_MountDuplicated(t *testing.T) {
	dupSandboxPath := []Pair[string, *types.Mount]{
		{Key: "/abc", Value: &types.Mount{Target: "/unique1"}},
		{Key: "/abc/", Value: &types.Mount{Target: "/unique2"}},
	}
	_, err := NewContainerRuntime(nil, WithMounts(dupSandboxPath))
	assert.ErrorContains(t, err, "found duplicate sandbox mount at")

	dupContainerPath := []Pair[string, *types.Mount]{
		{Key: "/unique1", Value: &types.Mount{Target: "/abc"}},
		{Key: "/unique2", Value: &types.Mount{Target: "/abc/"}},
	}
	_, err = NewContainerRuntime(nil, WithMounts(dupContainerPath))
	assert.ErrorContains(t, err, "found duplicate container mount at")
}

var mounts = []Pair[string, *types.Mount]{
	{Key: "/path/to/", Value: &types.Mount{
		Type:   "bind",
		Source: "/does/not/matter",
		Target: "/mnt/third",
	}},
	{Key: "/path/to/bar", Value: &types.Mount{
		Type:   "bind",
		Source: "/does/not/matter",
		Target: "/mnt/second",
	}},
	{Key: "/a/new/path", Value: &types.Mount{
		Type:   "bind",
		Source: "/does/not/matter",
		Target: "/mnt/second/fourth",
	}},
	{Key: "/path/to/foo/", Value: &types.Mount{
		Type:   "bind",
		Source: "/does/not/matter",
		Target: "/mnt/second/first",
	}},
}

//goland:noinspection GoSnakeCaseUsage
func testNewContainerRuntime_MountSorted(t *testing.T) {
	r, err := NewContainerRuntime(nil, WithMounts(mounts))
	assert.Nil(t, err)

	rt := r.(*containerRuntime)
	assert.Equal(t, 4, len(rt.mounts))
	for i := 1; i < len(rt.mounts); i++ {
		assert.True(t, rt.mounts[i-1].Key > rt.mounts[i].Key)
	}
	assert.Equal(t, 4, len(rt.pathMap))
	for i := 1; i < len(rt.pathMap); i++ {
		assert.True(t, rt.pathMap[i-1][0] > rt.pathMap[i][0])
	}
}

func TestContainerTranslatePath(t *testing.T) {
	rt, err := NewContainerRuntime(nil, WithMounts(mounts))
	assert.Nil(t, err)

	t.Run("exact-match", func(t *testing.T) {
		m := map[string]string{
			"/mnt/second/first":  "/path/to/foo/",
			"/mnt/second/fourth": "/a/new/path",
			"/mnt/second":        "/path/to/bar",
			"/mnt/third":         "/path/to/",
		}
		for k, v := range m {
			sbPath, ok := rt.TranslatePath(k)
			assert.True(t, ok)
			assert.Equal(t, v, sbPath)

			sbPath, ok = rt.TranslatePath(k + "/")
			assert.True(t, ok)
			assert.Equal(t, v, sbPath)
		}
	})

	t.Run("subpath", func(t *testing.T) {
		m := map[string]string{
			"/mnt/second/first/xxx":  "/path/to/foo/xxx",
			"/mnt/second/foobar":     "/path/to/bar/foobar",
			"/mnt/second/f":          "/path/to/bar/f",
			"/mnt/second/firstttttt": "/path/to/bar/firstttttt",
			"/mnt/third/abcxyz":      "/path/to/abcxyz",
		}
		for k, v := range m {
			sbPath, ok := rt.TranslatePath(k)
			assert.True(t, ok)
			assert.Equal(t, v, sbPath)
		}
	})

	t.Run("not-match", func(t *testing.T) {
		p := []string{"/mnt/", "/", "/foobar"}
		for _, v := range p {
			_, ok := rt.TranslatePath(v)
			assert.False(t, ok)
		}
	})
}

func TestContainerRun(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	engine := mock_container.NewMockEngine(ctrl)

	labels := map[string]string{"label": "value"}
	workdir := "/path/to/workdir"
	network := "net01"
	rt, err := NewContainerRuntime(engine,
		WithLabels(labels),
		WithWorkDir(workdir),
		WithNetwork(network),
		WithMounts(mounts),
	)
	assert.NoError(t, err)

	image := "drassi.run/docker/image"
	env := map[string]string{
		"A_SANDBOX_PATH": "/path/to/foobar",
		"A_NORMAL_ENV":   "hello-world",
	}
	entrypoint := []string{"/path/to/entrypoint.sh"}
	cmd := []string{"--flag", "with", "some", "arg"}
	engine.EXPECT().ContainerRun(ctx, gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, spec *types.ContainerSpec, _ *container.RunOptions) (string, error) {
			assert.Equal(t, labels, spec.Labels)
			assert.Equal(t, workdir, spec.WorkingDir)
			assert.Equal(t, network, spec.Endpoints[0].Target)
			assert.Equal(t, image, spec.Image)
			assert.Equal(t, entrypoint, spec.Entrypoint)
			assert.Equal(t, cmd, spec.Command)
			assert.True(t, spec.AutoRemove)

			e := map[string]string{
				"A_NORMAL_ENV":   "hello-world",
				"A_SANDBOX_PATH": "/mnt/third/foobar",
			}
			assert.Equal(t, e, spec.Environment)

			expectedMounts := make(map[string]*types.Mount)
			for _, m := range mounts {
				v := m.Value
				expectedMounts[v.Target] = v
			}
			actualMounts := make(map[string]*types.Mount)
			for _, m := range spec.Mounts {
				actualMounts[m.Target] = m
			}
			assert.Equal(t, expectedMounts, actualMounts)

			return "container_id", nil
		})

	err = rt.Run(ctx, image, entrypoint, cmd, env, nil)
	assert.NoError(t, err)
}
