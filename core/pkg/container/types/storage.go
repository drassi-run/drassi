/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types

import "io/fs"

type ContainerStorage struct {
	Mounts         []*Mount
	VolumesFrom    []string
	StorageOpt     map[string]string
	ReadonlyRootfs bool
}

// Mount represents a mount (volume).
//   - [github.com/moby/moby/api/types/mount.Mount]
//   - [github.com/moby/moby/api/types/container.MountPoint]
//   - [github.com/compose-spec/compose-go/v2/types.ServiceVolumeConfig]
type Mount struct {
	Type     string // "bind", "volume", "tmpfs"
	Source   string
	Target   string
	ReadOnly bool

	BindOptions   *BindOptions
	VolumeOptions *VolumeOptions
	TmpfsOptions  *TmpfsOptions
}

// BindOptions defines options specific to mounts of type "bind".
//   - [github.com/moby/moby/api/types/mount.BindOptions]
//   - [github.com/compose-spec/compose-go/v2/types.ServiceVolumeBind]
type BindOptions struct {
	Propagation    string // [r]shared | [r]slave | [r]private (default=rprivate)
	Consistency    string // consistent | delegated | cached (default=consistent)
	Recursive      string // enabled | disabled | writable | readonly (default=enabled)
	CreateHostPath bool
}

// VolumeOptions represents the options for a mount of type "volume".
//   - [github.com/moby/moby/api/types/mount.VolumeOptions]
//   - [github.com/compose-spec/compose-go/v2/types.ServiceVolumeVolume]
type VolumeOptions struct {
	NoCopy  bool
	Labels  map[string]string
	SubPath string

	// Driver config for volume mount.
	// [github.com/moby/moby/api/types/mount.Driver]
	Driver  string
	Options map[string]string
}

// TmpfsOptions defines options specific to mounts of type "tmpfs".
//   - [github.com/moby/moby/api/types/mount.TmpfsOptions]
//   - [github.com/compose-spec/compose-go/v2/types.ServiceVolumeTmpfs]
type TmpfsOptions struct {
	Size    int64
	Mode    fs.FileMode
	Options [][]string
}

// https://github.com/moby/moby/blob/docker-v29.7.2/api/types/volume/create_request.go
// https://github.com/containers/podman/blob/v5.2.4/pkg/domain/entities/types/volumes.go#L8-L21
type VolumeSpec struct {
	Name   string
	Labels map[string]string

	Driver  string
	Options map[string]string
}
