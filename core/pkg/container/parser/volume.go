/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package parser

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"

	"drassi.run/core/pkg/container/types"
	"github.com/docker/cli/cli/compose/loader"
	"github.com/docker/go-units"
)

// ParseVolume parses user-provided volume definitions into types.Mount format
//   - [github.com/containers/podman/v5/pkg/specgen.GenVolumeMounts]
//   - [github.com/compose-spec/compose-go/v2/loader.ParseVolume]
func ParseVolume(v string) (*types.Mount, error) {
	parsed, err := loader.ParseVolume(v)
	if err != nil {
		return nil, err
	}
	mount := &types.Mount{
		Type:     parsed.Type,
		Source:   parsed.Source,
		Target:   parsed.Target,
		ReadOnly: parsed.ReadOnly,
	}
	if bind := parsed.Bind; bind != nil {
		mount.BindOptions = &types.BindOptions{
			Propagation: bind.Propagation,
			Consistency: parsed.Consistency,
		}
	}
	if volume := parsed.Volume; volume != nil {
		mount.VolumeOptions = &types.VolumeOptions{
			NoCopy: volume.NoCopy,
		}
	}
	if tmp := parsed.Tmpfs; tmp != nil {
		mount.TmpfsOptions = &types.TmpfsOptions{
			Size: tmp.Size,
		}
	}
	return mount, nil
}

// ParseTmpfs parses user-provided tmpfs definitions into types.Mount format
//   - https://github.com/containers/podman/blob/v5.2.5/pkg/specgenutil/volumes.go#L645
func ParseTmpfs(t string) (*types.Mount, error) {
	split := strings.Split(t, ":")
	target := split[0]
	if err := validateVolumeContainerDir(target); err != nil {
		return nil, err
	}
	mount := &types.Mount{
		Type:   "tmpfs",
		Target: target,
	}

	if len(split) > 1 {
		options := strings.Split(split[1], ",")
		mount.TmpfsOptions = &types.TmpfsOptions{}
		for _, opt := range options {
			k, v, _ := strings.Cut(opt, "=")
			k = strings.ToLower(k)
			switch k {
			case "size":
				if size, err := units.RAMInBytes(v); err != nil {
					return nil, err
				} else {
					mount.TmpfsOptions.Size = size
				}
			case "readonly", "ro":
				mount.ReadOnly = true
			case "readwrite", "rw":
				mount.ReadOnly = false
			case "mode":
				if ui64, err := strconv.ParseUint(v, 8, 32); err != nil {
					return nil, err
				} else {
					mount.TmpfsOptions.Mode = fs.FileMode(ui64)
				}
			default:
				o := []string{k}
				if v != "" {
					o = append(o, v)
				}
				mount.TmpfsOptions.Options = append(mount.TmpfsOptions.Options, o)
			}
		}
	}
	return mount, nil
}

// validateVolumeContainerDir validates a volume mount's destination directory.
func validateVolumeContainerDir(path string) error {
	if path == "" {
		return errors.New("container directory cannot be empty")
	}
	if !filepath.IsAbs(path) {
		return fmt.Errorf("invalid container path %q, must be an absolute path", path)
	}
	return nil
}
