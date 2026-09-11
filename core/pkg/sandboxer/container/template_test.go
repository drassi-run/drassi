/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package container

import (
	"bytes"
	"testing"

	"drassi.run/core/pkg/container/types"
	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/require"
)

func TestTemplateUnmarshalTOML(t *testing.T) {
	t.Run("unmarshals template correctly", func(t *testing.T) {
		tomlData := []byte(`
image = "ghcr.io/drassi-run/ubuntu:26.04"
network_mode = "bridge"
privileged = true
user = "1000:1000"
working_dir = "/app"
command = ["run", "app"]
entrypoint = ["/entrypoint.sh"]
environment = { FOO = "bar", BAZ = "qux" }
labels = { "org.drassi.env" = "test" }
annotations = { "note" = "template" }
volumes = [
    "/host/path:/container/path:ro",
    "/data:/data"
]
ports = [
    "8080:80/tcp",
    "127.0.0.1:9090:90"
]
devices = ["/dev/kvm"]
device_cgroup_rules = ["c 10:200 rwm"]
cap_add = ["SYS_ADMIN"]
cap_drop = ["NET_RAW"]
security_opt = ["no-new-privileges:true"]
sysctls = { "net.ipv4.ip_forward" = "1" }
volumes_from = ["other-container:ro"]
storage_opt = { "size" = "20G" }
group_add = ["wheel"]
cpus = "1.5"
memory = 1073741824
`)

		var tmpl Template
		err := tmpl.UnmarshalTOML(tomlData)
		require.NoError(t, err)

		require.Equal(t, "ghcr.io/drassi-run/ubuntu:26.04", tmpl.Image)
		require.Equal(t, "bridge", tmpl.NetworkMode)
		require.True(t, tmpl.Privileged)
		require.Equal(t, "1000:1000", tmpl.User)
		require.Equal(t, "/app", tmpl.WorkingDir)
		require.Equal(t, []string{"run", "app"}, tmpl.Command)
		require.Equal(t, []string{"/entrypoint.sh"}, tmpl.Entrypoint)
		require.Equal(t, map[string]string{"FOO": "bar", "BAZ": "qux"}, tmpl.Environment)
		require.Equal(t, map[string]string{"org.drassi.env": "test"}, tmpl.Labels)
		require.Equal(t, map[string]string{"note": "template"}, tmpl.Annotations)
		require.Len(t, tmpl.Mounts, 2)
		require.Equal(t, "/host/path", tmpl.Mounts[0].Source)
		require.Equal(t, "/container/path", tmpl.Mounts[0].Target)
		require.True(t, tmpl.Mounts[0].ReadOnly)
		require.Equal(t, "/data", tmpl.Mounts[1].Source)
		require.Equal(t, "/data", tmpl.Mounts[1].Target)
		require.False(t, tmpl.Mounts[1].ReadOnly)

		require.Len(t, tmpl.Publish, 2)
		require.Equal(t, uint16(8080), tmpl.Publish[0].HostPort)
		require.Equal(t, uint16(80), tmpl.Publish[0].ContainerPort)
		require.Equal(t, "tcp", tmpl.Publish[0].Protocol)
		require.Equal(t, "127.0.0.1", tmpl.Publish[1].HostIP)
		require.Equal(t, uint16(9090), tmpl.Publish[1].HostPort)
		require.Equal(t, uint16(90), tmpl.Publish[1].ContainerPort)

		require.Equal(t, []string{"/dev/kvm"}, tmpl.Devices)
		require.Equal(t, []string{"c 10:200 rwm"}, tmpl.DeviceCgroupRules)
		require.Equal(t, []string{"SYS_ADMIN"}, tmpl.CapAdd)
		require.Equal(t, []string{"NET_RAW"}, tmpl.CapDrop)
		require.Equal(t, []string{"no-new-privileges:true"}, tmpl.SecurityOpt)
		require.Equal(t, map[string]string{"net.ipv4.ip_forward": "1"}, tmpl.Sysctls)
		require.Equal(t, []string{"other-container:ro"}, tmpl.VolumesFrom)
		require.Equal(t, map[string]string{"size": "20G"}, tmpl.StorageOpt)
		require.Equal(t, []string{"wheel"}, tmpl.GroupAdd)
		require.Equal(t, "1.5", tmpl.CPUS)
		require.Equal(t, int64(1073741824), tmpl.Memory)
	})

	t.Run("decodes via toml.Decoder with EnableUnmarshalerInterface", func(t *testing.T) {
		type Wrapper struct {
			Template *Template `toml:"template"`
		}
		raw := []byte(`
[template]
image = "alpine:3.20"
network_mode = "none"
volumes = ["/tmp:/tmp:ro"]
`)
		var w Wrapper
		err := toml.NewDecoder(bytes.NewReader(raw)).EnableUnmarshalerInterface().Decode(&w)
		require.NoError(t, err)
		require.NotNil(t, w.Template)
		require.Equal(t, "alpine:3.20", w.Template.Image)
		require.Equal(t, "none", w.Template.NetworkMode)
		require.Len(t, w.Template.Mounts, 1)
		require.Equal(t, "/tmp", w.Template.Mounts[0].Source)
		require.Equal(t, "/tmp", w.Template.Mounts[0].Target)
		require.True(t, w.Template.Mounts[0].ReadOnly)
	})

	t.Run("invalid toml returns error", func(t *testing.T) {
		var tmpl Template
		err := tmpl.UnmarshalTOML([]byte("invalid toml = ="))
		require.Error(t, err)
	})
}

func TestTemplateCopy(t *testing.T) {
	t.Run("nil receiver returns nil", func(t *testing.T) {
		var tmpl *Template
		require.Nil(t, tmpl.Copy())
		require.Nil(t, tmpl.Clone())
		require.Nil(t, tmpl.ContainerSpec())
	})

	t.Run("deep copy preserves all fields and isolates mutations", func(t *testing.T) {
		orig := &Template{
			Image:             "ubuntu:latest",
			NetworkMode:       "bridge",
			Command:           []string{"echo", "hello"},
			Entrypoint:        []string{"/bin/sh"},
			Environment:       map[string]string{"A": "1"},
			Labels:            map[string]string{"L": "2"},
			Annotations:       map[string]string{"N": "3"},
			Devices:           []string{"/dev/null"},
			DeviceCgroupRules: []string{"rwm"},
			Exposes: []*types.Port{
				{Number: 80, Protocol: "tcp"},
			},
			Publish: []*types.PortBinding{
				{HostPort: 8080, ContainerPort: 80, Protocol: "tcp"},
			},
			Mounts: []*types.Mount{
				{Type: "bind", Source: "/host", Target: "/container", ReadOnly: true},
			},
			VolumesFrom: []string{"base"},
			StorageOpt:  map[string]string{"size": "10G"},
			GroupAdd:    []string{"docker"},
			CapAdd:      []string{"NET_ADMIN"},
			CapDrop:     []string{"SYS_CHROOT"},
			SecurityOpt: []string{"apparmor=unconfined"},
			Sysctls:     map[string]string{"net.ipv4.tcp_syncookies": "1"},
		}

		cp := orig.Copy()
		require.NotNil(t, cp)
		require.NotSame(t, orig, cp)

		// Mutate original slices and maps
		orig.Command[0] = "sleep"
		orig.Entrypoint[0] = "/bin/bash"
		orig.Environment["A"] = "modified"
		orig.Labels["L"] = "modified"
		orig.Annotations["N"] = "modified"
		orig.Devices[0] = "/dev/zero"
		orig.DeviceCgroupRules[0] = "none"
		orig.Exposes[0].Number = 443
		orig.Publish[0].HostPort = 8443
		orig.Mounts[0].Source = "/other"
		orig.VolumesFrom[0] = "modified"
		orig.StorageOpt["size"] = "50G"
		orig.GroupAdd[0] = "root"
		orig.CapAdd[0] = "ALL"
		orig.CapDrop[0] = "ALL"
		orig.SecurityOpt[0] = "seccomp=unconfined"
		orig.Sysctls["net.ipv4.tcp_syncookies"] = "0"

		// Verify copied values were isolated
		require.Equal(t, "echo", cp.Command[0])
		require.Equal(t, "/bin/sh", cp.Entrypoint[0])
		require.Equal(t, "1", cp.Environment["A"])
		require.Equal(t, "2", cp.Labels["L"])
		require.Equal(t, "3", cp.Annotations["N"])
		require.Equal(t, "/dev/null", cp.Devices[0])
		require.Equal(t, "rwm", cp.DeviceCgroupRules[0])
		require.Equal(t, uint16(80), cp.Exposes[0].Number)
		require.Equal(t, uint16(8080), cp.Publish[0].HostPort)
		require.Equal(t, "/host", cp.Mounts[0].Source)
		require.Equal(t, "base", cp.VolumesFrom[0])
		require.Equal(t, "10G", cp.StorageOpt["size"])
		require.Equal(t, "docker", cp.GroupAdd[0])
		require.Equal(t, "NET_ADMIN", cp.CapAdd[0])
		require.Equal(t, "SYS_CHROOT", cp.CapDrop[0])
		require.Equal(t, "apparmor=unconfined", cp.SecurityOpt[0])
		require.Equal(t, "1", cp.Sysctls["net.ipv4.tcp_syncookies"])

		// ContainerSpec() returns *types.ContainerSpec
		spec := orig.ContainerSpec()
		require.NotNil(t, spec)
		require.Equal(t, orig.Image, spec.Image)
	})
}
