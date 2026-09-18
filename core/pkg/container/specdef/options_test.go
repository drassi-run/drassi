/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package specdef

import (
	"errors"
	"testing"

	mock_container "drassi.run/core/mock/container"
	"drassi.run/core/pkg/container/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestApply(t *testing.T) {
	t.Run("error propagation", func(t *testing.T) {
		spec := &types.ContainerSpec{}
		err := Apply(spec,
			SetWorkdir("/app"),
			func(s *types.ContainerSpec) error {
				return errors.New("boom")
			},
			SetWorkdir("/ignored"),
		)
		require.Error(t, err)
		assert.Equal(t, "boom", err.Error())
		assert.Equal(t, "/app", spec.WorkingDir)
	})

	t.Run("happy path", func(t *testing.T) {
		s := &types.ContainerSpec{}
		err := Apply(s,
			SetWorkdir("/app"),
			SetCmd([]string{"run"}, nil),
		)
		require.NoError(t, err)
		assert.Equal(t, "/app", s.WorkingDir)
		assert.Equal(t, []string{"run"}, s.Entrypoint)
	})
}

func TestSetLabels(t *testing.T) {
	t.Run("set labels on spec and mounts", func(t *testing.T) {
		spec := &types.ContainerSpec{
			Mounts: []*types.Mount{
				{Type: "volume", Source: "v1", Target: "/v1"},
				{Type: "bind", Source: "/host", Target: "/mnt"},
			},
		}
		err := SetLabels(map[string]string{"foo": "bar"})(spec)
		require.NoError(t, err)
		assert.Equal(t, "bar", spec.Labels["foo"])
		require.NotNil(t, spec.Mounts[0].VolumeOptions)
		assert.Equal(t, "bar", spec.Mounts[0].VolumeOptions.Labels["foo"])
		assert.Nil(t, spec.Mounts[1].VolumeOptions)
	})

	t.Run("empty or nil labels is noop", func(t *testing.T) {
		s := &types.ContainerSpec{
			Mounts: []*types.Mount{{Type: "volume"}},
		}
		require.NoError(t, SetLabels(nil)(s))
		require.NoError(t, SetLabels(map[string]string{})(s))
		assert.Nil(t, s.Labels)
		assert.Nil(t, s.Mounts[0].VolumeOptions)
	})

	t.Run("merge with existing labels and handles nil mount", func(t *testing.T) {
		s := &types.ContainerSpec{
			Labels: map[string]string{"k1": "v1"},
			Mounts: []*types.Mount{
				nil,
				{
					Type: "volume",
					VolumeOptions: &types.VolumeOptions{
						Labels: map[string]string{"existing": "true"},
					},
				},
			},
		}
		require.NoError(t, SetLabels(map[string]string{"k2": "v2"})(s))
		assert.Equal(t, "v1", s.Labels["k1"])
		assert.Equal(t, "v2", s.Labels["k2"])
		assert.Equal(t, "true", s.Mounts[1].VolumeOptions.Labels["existing"])
		assert.Equal(t, "v2", s.Mounts[1].VolumeOptions.Labels["k2"])
	})
}

func TestSetCmd(t *testing.T) {
	spec := &types.ContainerSpec{
		Command: []string{"old"},
	}
	err := SetCmd([]string{"entry"}, []string{"cmd1", "cmd2"})(spec)
	require.NoError(t, err)
	assert.Equal(t, []string{"entry"}, spec.Entrypoint)
	assert.Equal(t, []string{"cmd1", "cmd2"}, spec.Command)

	// Entrypoint only clears command
	err = SetCmd([]string{"newentry"}, nil)(spec)
	require.NoError(t, err)
	assert.Equal(t, []string{"newentry"}, spec.Entrypoint)
	assert.Nil(t, spec.Command)
}

func TestSetNetwork(t *testing.T) {
	t.Run("empty endpoints", func(t *testing.T) {
		spec := &types.ContainerSpec{}
		err := SetNetwork("net1")(spec)
		require.NoError(t, err)
		require.Len(t, spec.Endpoints, 1)
		assert.Equal(t, "net1", spec.Endpoints[0].Target)
	})

	t.Run("single default endpoint", func(t *testing.T) {
		spec := &types.ContainerSpec{
			Endpoints: []*types.Endpoint{{Target: ""}},
		}
		err := SetNetwork("net2")(spec)
		require.NoError(t, err)
		require.Len(t, spec.Endpoints, 1)
		assert.Equal(t, "net2", spec.Endpoints[0].Target)
	})

	t.Run("conflicting network error", func(t *testing.T) {
		spec := &types.ContainerSpec{
			Endpoints: []*types.Endpoint{{Target: "existing"}},
		}
		err := SetNetwork("net3")(spec)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can't overwrite non-default network")
	})

	t.Run("multiple endpoints error", func(t *testing.T) {
		spec := &types.ContainerSpec{
			Endpoints: []*types.Endpoint{{Target: "a"}, {Target: "b"}},
		}
		err := SetNetwork("net3")(spec)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "only one network per container")
	})
}

func TestAddMount(t *testing.T) {
	spec := &types.ContainerSpec{}
	m1 := &types.Mount{Type: "bind", Source: "/a", Target: "/b"}
	err := AddMount(m1)(spec)
	require.NoError(t, err)
	assert.Equal(t, []*types.Mount{m1}, spec.Mounts)
}

func TestMountVolume(t *testing.T) {
	spec := &types.ContainerSpec{}
	err := MountVolume("myvol", "/data")(spec)
	require.NoError(t, err)
	require.Len(t, spec.Mounts, 1)
	assert.Equal(t, "volume", spec.Mounts[0].Type)
	assert.Equal(t, "myvol", spec.Mounts[0].Source)
	assert.Equal(t, "/data", spec.Mounts[0].Target)
}

func TestMountApiSocket(t *testing.T) {
	ctrl := gomock.NewController(t)

	t.Run("unix socket", func(t *testing.T) {
		spec := &types.ContainerSpec{}
		engine := mock_container.NewMockEngine(ctrl)
		engine.EXPECT().Address().Return("unix:///var/run/docker.sock")
		err := MountApiSocket(engine)(spec)
		require.NoError(t, err)
		require.Len(t, spec.Mounts, 1)
		assert.Equal(t, "bind", spec.Mounts[0].Type)
		assert.Equal(t, "/var/run/docker.sock", spec.Mounts[0].Source)
		assert.Equal(t, "/var/run/docker.sock", spec.Mounts[0].Target)
	})

	t.Run("non-unix socket is noop", func(t *testing.T) {
		spec := &types.ContainerSpec{}
		engine := mock_container.NewMockEngine(ctrl)
		engine.EXPECT().Address().Return("tcp://127.0.0.1:2375")
		err := MountApiSocket(engine)(spec)
		require.NoError(t, err)
		assert.Empty(t, spec.Mounts)
	})

	t.Run("nil engine is noop", func(t *testing.T) {
		spec := &types.ContainerSpec{}
		err := MountApiSocket(nil)(spec)
		require.NoError(t, err)
		assert.Empty(t, spec.Mounts)
	})
}

func TestSetWorkdir(t *testing.T) {
	spec := &types.ContainerSpec{}
	err := SetWorkdir("/workspace")(spec)
	require.NoError(t, err)
	assert.Equal(t, "/workspace", spec.WorkingDir)

	err = SetWorkdir("/another")(spec)
	require.NoError(t, err)
	assert.Equal(t, "/another", spec.WorkingDir)
}

func TestSetCIEnv(t *testing.T) {
	t.Run("nil environment", func(t *testing.T) {
		spec := &types.ContainerSpec{}
		err := SetCIEnv()(spec)
		require.NoError(t, err)
		assert.Equal(t, "true", spec.Environment["CI"])
		assert.Equal(t, "true", spec.Environment["GITHUB_ACTIONS"])
	})

	t.Run("existing environment", func(t *testing.T) {
		spec := &types.ContainerSpec{
			Environment: map[string]string{"EXISTING": "value"},
		}
		err := SetCIEnv()(spec)
		require.NoError(t, err)
		assert.Equal(t, "value", spec.Environment["EXISTING"])
		assert.Equal(t, "true", spec.Environment["CI"])
		assert.Equal(t, "true", spec.Environment["GITHUB_ACTIONS"])
	})
}
