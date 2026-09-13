/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
)

type ContainerStorage struct {
	Mounts         []*Mount `json:"volumes,omitempty"`
	VolumesFrom    []string `json:"volumes_from,omitempty"`
	StorageOpt     Mapping  `json:"storage_opt,omitempty"`
	ReadonlyRootfs bool     `json:"readonly_rootfs,omitempty"`
}

// Mount represents a mount (volume).
//   - [github.com/moby/moby/api/types/mount.Mount]
//   - [github.com/moby/moby/api/types/container.MountPoint]
//   - [github.com/compose-spec/compose-go/v2/types.ServiceVolumeConfig]
type Mount struct {
	Type     string `json:"type,omitempty"` // "bind", "volume", "image", "tmpfs"
	Source   string `json:"source,omitempty"`
	Target   string `json:"target,omitempty"`
	ReadOnly bool   `json:"read_only,omitempty"`

	BindOptions   *BindOptions   `json:"bind,omitempty"`
	VolumeOptions *VolumeOptions `json:"volume,omitempty"`
	ImageOptions  *ImageOptions  `json:"image,omitempty"`
	TmpfsOptions  *TmpfsOptions  `json:"tmpfs,omitempty"`
}

var ParseVolume func(v string) (*Mount, error)

func (m *Mount) UnmarshalJSONFrom(d *jsontext.Decoder) error {
	switch k := d.PeekKind(); k {
	case jsontext.KindString:
		var s string
		if err := json.UnmarshalDecode(d, &s); err != nil {
			return err
		}
		if ParseVolume == nil {
			return errors.New("types: volume parser not registered (import _ \"drassi.run/core/pkg/container/parser\")")
		}
		parsed, err := ParseVolume(s)
		if err != nil {
			return err
		}
		*m = *parsed
		return nil
	case jsontext.KindBeginObject:
		type alias Mount
		return json.UnmarshalDecode(d, (*alias)(m))
	default:
		return fmt.Errorf("expected string or object for Mount, got %v", k)
	}
}

// BindOptions defines options specific to mounts of type "bind".
//   - [github.com/moby/moby/api/types/mount.BindOptions]
//   - [github.com/compose-spec/compose-go/v2/types.ServiceVolumeBind]
type BindOptions struct {
	Propagation    string `json:"propagation,omitempty"` // [r]shared | [r]slave | [r]private (default=rprivate)
	Consistency    string `json:"consistency,omitempty"` // consistent | delegated | cached (default=consistent)
	Recursive      string `json:"recursive,omitempty"`   // enabled | disabled | writable | readonly (default=enabled)
	CreateHostPath bool   `json:"create_host_path,omitempty"`
}

// VolumeOptions represents the options for a mount of type "volume".
//   - [github.com/moby/moby/api/types/mount.VolumeOptions]
//   - [github.com/compose-spec/compose-go/v2/types.ServiceVolumeVolume]
type VolumeOptions struct {
	NoCopy  bool              `json:"no_copy,omitempty"`
	Labels  map[string]string `json:"labels,omitempty"`
	SubPath string            `json:"subpath,omitempty"`

	// Driver config for volume mount.
	// [github.com/moby/moby/api/types/mount.Driver]
	Driver  string            `json:"driver,omitempty"`
	Options map[string]string `json:"options,omitempty"`
}

// ImageOptions represents the options for a mount of type "image".
//   - [github.com/moby/moby/api/types/mount.ImageOptions]
//   - [github.com/compose-spec/compose-go/v2/types.ServiceVolumeImage]
type ImageOptions struct {
	Subpath string `json:"subpath,omitempty"`
}

// TmpfsOptions defines options specific to mounts of type "tmpfs".
//   - [github.com/moby/moby/api/types/mount.TmpfsOptions]
//   - [github.com/compose-spec/compose-go/v2/types.ServiceVolumeTmpfs]
type TmpfsOptions struct {
	Size    int64       `json:"size,omitempty"`
	Mode    fs.FileMode `json:"mode,omitempty"`
	Options [][]string  `json:"options,omitempty"`
}

// https://github.com/moby/moby/blob/docker-v29.7.2/api/types/volume/create_request.go
// https://github.com/containers/podman/blob/v5.2.4/pkg/domain/entities/types/volumes.go#L8-L21
type VolumeSpec struct {
	Name   string            `json:"name,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`

	Driver  string            `json:"driver,omitempty"`
	Options map[string]string `json:"options,omitempty"`
}
