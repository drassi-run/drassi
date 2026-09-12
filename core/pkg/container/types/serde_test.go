/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types_test

import (
	"encoding/json/v2"
	"testing"

	"drassi.run/core/pkg/container/types"
	"github.com/stretchr/testify/require"
)

func TestMountUnmarshalJSON(t *testing.T) {
	t.Run("shorthand string volume", func(t *testing.T) {
		jsonStr := `"/var/run/docker.sock:/var/run/docker.sock:ro"`
		var m types.Mount
		err := json.Unmarshal([]byte(jsonStr), &m)
		require.NoError(t, err)
		require.Equal(t, "bind", m.Type)
		require.Equal(t, "/var/run/docker.sock", m.Source)
		require.Equal(t, "/var/run/docker.sock", m.Target)
		require.True(t, m.ReadOnly)
	})

	t.Run("object volume", func(t *testing.T) {
		jsonStr := `{"type": "volume", "source": "my-vol", "target": "/data", "read_only": false}`
		var m types.Mount
		err := json.Unmarshal([]byte(jsonStr), &m)
		require.NoError(t, err)
		require.Equal(t, "volume", m.Type)
		require.Equal(t, "my-vol", m.Source)
		require.Equal(t, "/data", m.Target)
		require.False(t, m.ReadOnly)
	})
}

func TestPortBindingUnmarshalJSON(t *testing.T) {
	t.Run("shorthand string port without host ip", func(t *testing.T) {
		jsonStr := `"8080:80"`
		var pb types.PortBinding
		err := json.Unmarshal([]byte(jsonStr), &pb)
		require.NoError(t, err)
		require.Equal(t, uint16(8080), pb.HostPort)
		require.Equal(t, uint16(80), pb.ContainerPort)
		require.Equal(t, "tcp", pb.Protocol)
	})

	t.Run("shorthand string port with host ip and protocol", func(t *testing.T) {
		jsonStr := `"127.0.0.1:3000:3000/udp"`
		var pb types.PortBinding
		err := json.Unmarshal([]byte(jsonStr), &pb)
		require.NoError(t, err)
		require.Equal(t, "127.0.0.1", pb.HostIP)
		require.Equal(t, uint16(3000), pb.HostPort)
		require.Equal(t, uint16(3000), pb.ContainerPort)
		require.Equal(t, "udp", pb.Protocol)
	})

	t.Run("object port binding", func(t *testing.T) {
		jsonStr := `{"host_ip": "0.0.0.0", "host_port": 9000, "container_port": 9000, "protocol": "tcp"}`
		var pb types.PortBinding
		err := json.Unmarshal([]byte(jsonStr), &pb)
		require.NoError(t, err)
		require.Equal(t, "0.0.0.0", pb.HostIP)
		require.Equal(t, uint16(9000), pb.HostPort)
		require.Equal(t, uint16(9000), pb.ContainerPort)
		require.Equal(t, "tcp", pb.Protocol)
	})
}

func TestContainerSpecUnmarshalJSON(t *testing.T) {
	raw := `{
		"image": "ghcr.io/drassi-run/ubuntu:26.04",
		"working_dir": "/workspace",
		"environment": {"FOO": "bar", "DEBUG": "true"},
		"volumes": [
			"/var/run/docker.sock:/var/run/docker.sock",
			"/data:/data:ro"
		],
		"ports": [
			"8080:80"
		],
		"network_mode": "host",
		"privileged": true,
		"user": "1000:1000",
		"cap_add": ["SYS_ADMIN"],
		"cap_drop": ["NET_RAW"],
		"devices": ["/dev/kvm"],
		"cpus": "2",
		"memory": 2147483648,
		"shm_size": 1073741824
	}`

	var spec types.ContainerSpec
	err := json.Unmarshal([]byte(raw), &spec)
	require.NoError(t, err)

	require.Equal(t, "ghcr.io/drassi-run/ubuntu:26.04", spec.Image)
	require.Equal(t, "/workspace", spec.WorkingDir)
	require.Equal(t, "bar", spec.Environment["FOO"])
	require.Equal(t, "true", spec.Environment["DEBUG"])

	// Volumes
	require.Len(t, spec.Mounts, 2)
	require.Equal(t, "/var/run/docker.sock", spec.Mounts[0].Source)
	require.Equal(t, "/var/run/docker.sock", spec.Mounts[0].Target)
	require.False(t, spec.Mounts[0].ReadOnly)
	require.Equal(t, "/data", spec.Mounts[1].Source)
	require.Equal(t, "/data", spec.Mounts[1].Target)
	require.True(t, spec.Mounts[1].ReadOnly)

	// Ports
	require.Len(t, spec.Publish, 1)
	require.Equal(t, uint16(8080), spec.Publish[0].HostPort)
	require.Equal(t, uint16(80), spec.Publish[0].ContainerPort)

	// Security
	require.Equal(t, "host", spec.NetworkMode)
	require.True(t, spec.Privileged)
	require.Equal(t, "1000:1000", spec.User)
	require.Equal(t, []string{"SYS_ADMIN"}, spec.CapAdd)
	require.Equal(t, []string{"NET_RAW"}, spec.CapDrop)

	// Devices & Resources
	require.Equal(t, []string{"/dev/kvm"}, spec.Devices)
	require.Equal(t, "2", spec.CPUS)
	require.Equal(t, int64(2147483648), spec.Memory)
	require.Equal(t, int64(1073741824), spec.ShmSize)
}
