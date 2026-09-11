/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types_test

import (
	"encoding/json/v2"
	"testing"

	_ "drassi.run/core/pkg/container/parser"
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
	require.Equal(t, types.UnitBytes(2147483648), spec.Memory)
	require.Equal(t, types.UnitBytes(1073741824), spec.ShmSize)
}

func TestMappingUnmarshalJSON(t *testing.T) {
	t.Run("object mapping with mixed values", func(t *testing.T) {
		raw := `{"FOO": "bar", "PORT": 8080, "BOOL": true, "EMPTY": null}`
		var m types.Mapping
		err := json.Unmarshal([]byte(raw), &m)
		require.NoError(t, err)
		require.Equal(t, "bar", m["FOO"])
		require.Equal(t, "8080", m["PORT"])
		require.Equal(t, "true", m["BOOL"])
		require.Equal(t, "", m["EMPTY"])
	})

	t.Run("list of key-value and key-only strings", func(t *testing.T) {
		raw := `["FOO=bar", "BAZ=qux=extra", "FLAG", "EMPTY="]`
		var m types.Mapping
		err := json.Unmarshal([]byte(raw), &m)
		require.NoError(t, err)
		require.Equal(t, "bar", m["FOO"])
		require.Equal(t, "qux=extra", m["BAZ"])
		require.Equal(t, "", m["FLAG"])
		require.Equal(t, "", m["EMPTY"])
	})

	t.Run("null mapping", func(t *testing.T) {
		var m types.Mapping = types.Mapping{"A": "1"}
		err := json.Unmarshal([]byte("null"), &m)
		require.NoError(t, err)
		require.Nil(t, m)
	})
}

func TestPortUnmarshalJSON(t *testing.T) {
	t.Run("port from string number", func(t *testing.T) {
		var p types.Port
		err := json.Unmarshal([]byte(`"80"`), &p)
		require.NoError(t, err)
		require.Equal(t, uint16(80), p.Number)
		require.Equal(t, "tcp", p.Protocol)
		require.Equal(t, "80/tcp", p.String())
	})

	t.Run("port from string with protocol", func(t *testing.T) {
		var p types.Port
		err := json.Unmarshal([]byte(`"53/udp"`), &p)
		require.NoError(t, err)
		require.Equal(t, uint16(53), p.Number)
		require.Equal(t, "udp", p.Protocol)
		require.Equal(t, "53/udp", p.String())
	})

	t.Run("port from integer number", func(t *testing.T) {
		var p types.Port
		err := json.Unmarshal([]byte(`443`), &p)
		require.NoError(t, err)
		require.Equal(t, uint16(443), p.Number)
		require.Equal(t, "tcp", p.Protocol)
	})

	t.Run("port from object", func(t *testing.T) {
		var p types.Port
		err := json.Unmarshal([]byte(`{"number": 8080, "protocol": "tcp"}`), &p)
		require.NoError(t, err)
		require.Equal(t, uint16(8080), p.Number)
		require.Equal(t, "tcp", p.Protocol)
	})
}

func TestUnitBytesUnmarshalJSON(t *testing.T) {
	t.Run("from integer number", func(t *testing.T) {
		var u types.UnitBytes
		err := json.Unmarshal([]byte(`2147483648`), &u)
		require.NoError(t, err)
		require.Equal(t, types.UnitBytes(2147483648), u)
		require.Equal(t, int64(2147483648), int64(u))
	})

	t.Run("from numeric string", func(t *testing.T) {
		var u types.UnitBytes
		err := json.Unmarshal([]byte(`"2147483648"`), &u)
		require.NoError(t, err)
		require.Equal(t, types.UnitBytes(2147483648), u)
	})

	t.Run("from human-readable unit strings", func(t *testing.T) {
		cases := []struct {
			input    string
			expected types.UnitBytes
		}{
			{`"2g"`, 2 * 1024 * 1024 * 1024},
			{`"512m"`, 512 * 1024 * 1024},
			{`"64mb"`, 64 * 1024 * 1024},
			{`"1k"`, 1024},
		}

		for _, tc := range cases {
			var u types.UnitBytes
			err := json.Unmarshal([]byte(tc.input), &u)
			require.NoError(t, err)
			require.Equal(t, tc.expected, u)
		}
	})

	t.Run("invalid string returns error", func(t *testing.T) {
		var u types.UnitBytes
		err := json.Unmarshal([]byte(`"invalid-size"`), &u)
		require.Error(t, err)
	})
}

func TestContainerSpecShortFormUnmarshalJSON(t *testing.T) {
	raw := `{
		"image": "ghcr.io/drassi-run/ubuntu:26.04",
		"environment": [
			"FOO=bar",
			"DEBUG=true",
			"FLAG"
		],
		"labels": [
			"app=drassi",
			"env=prod"
		],
		"annotations": [
			"note=ready"
		],
		"sysctls": [
			"net.ipv4.ip_forward=1"
		],
		"storage_opt": [
			"size=20G"
		],
		"expose": [
			"80",
			"53/udp",
			443
		],
		"memory": "2g",
		"mem_reservation": "1g",
		"shm_size": "64m"
	}`

	var spec types.ContainerSpec
	err := json.Unmarshal([]byte(raw), &spec)
	require.NoError(t, err)

	require.Equal(t, "bar", spec.Environment["FOO"])
	require.Equal(t, "true", spec.Environment["DEBUG"])
	require.Equal(t, "", spec.Environment["FLAG"])

	require.Equal(t, "drassi", spec.Labels["app"])
	require.Equal(t, "prod", spec.Labels["env"])
	require.Equal(t, "ready", spec.Annotations["note"])
	require.Equal(t, "1", spec.Sysctls["net.ipv4.ip_forward"])
	require.Equal(t, "20G", spec.StorageOpt["size"])

	require.Len(t, spec.Exposes, 3)
	require.Equal(t, uint16(80), spec.Exposes[0].Number)
	require.Equal(t, "tcp", spec.Exposes[0].Protocol)
	require.Equal(t, uint16(53), spec.Exposes[1].Number)
	require.Equal(t, "udp", spec.Exposes[1].Protocol)
	require.Equal(t, uint16(443), spec.Exposes[2].Number)
	require.Equal(t, "tcp", spec.Exposes[2].Protocol)

	require.Equal(t, types.UnitBytes(2*1024*1024*1024), spec.Memory)
	require.Equal(t, types.UnitBytes(1024*1024*1024), spec.MemReservation)
	require.Equal(t, types.UnitBytes(64*1024*1024), spec.ShmSize)
}
