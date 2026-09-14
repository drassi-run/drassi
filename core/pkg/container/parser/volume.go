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

	"unicode"

	"drassi.run/core/pkg/container/types"
	"github.com/docker/go-units"
)

const endOfSpec = rune(0)

// ParseVolume parses user-provided volume definitions into types.Mount format
//   - [github.com/containers/podman/v5/pkg/specgen.GenVolumeMounts]
//   - [github.com/compose-spec/compose-go/v2/loader.ParseVolume]
func ParseVolume(spec string) (*types.Mount, error) {
	if len(spec) == 0 {
		return nil, errors.New("invalid empty volume spec")
	}

	mount := new(types.Mount)
	buffer := make([]rune, 0, len(spec))
	for _, char := range spec + string(endOfSpec) {
		switch {
		case isWindowsDrive(buffer, char):
			buffer = append(buffer, char)
		case char == ':' || char == endOfSpec:
			if err := populateFieldFromBuffer(char, buffer, mount); err != nil {
				populateType(mount)
				return nil, fmt.Errorf("invalid spec: %s: %w", spec, err)
			}
			buffer = buffer[:0] // reset, but reuse capacity
		default:
			buffer = append(buffer, char)
		}
	}

	populateType(mount)
	return mount, nil
}

func isWindowsDrive(buffer []rune, char rune) bool {
	return char == ':' && len(buffer) == 1 && unicode.IsLetter(buffer[0])
}

func populateFieldFromBuffer(char rune, buffer []rune, mount *types.Mount) error {
	strBuffer := string(buffer)
	switch {
	case len(buffer) == 0:
		return errors.New("empty section between colons")
	// Anonymous volume
	case mount.Source == "" && char == endOfSpec:
		mount.Target = strBuffer
		return nil
	case mount.Source == "":
		mount.Source = strBuffer
		return nil
	case mount.Target == "":
		mount.Target = strBuffer
		return nil
	case char == ':':
		return errors.New("too many colons")
	}
	for option := range strings.SplitSeq(strBuffer, ",") {
		switch option {
		case "ro":
			mount.ReadOnly = true
		case "rw":
			mount.ReadOnly = false
		case "nocopy":
			if mount.VolumeOptions == nil {
				mount.VolumeOptions = new(types.VolumeOptions)
			}
			mount.VolumeOptions.NoCopy = true
		case "rprivate", "private", "rshared", "shared", "rslave", "slave":
			if mount.BindOptions == nil {
				mount.BindOptions = new(types.BindOptions)
			}
			mount.BindOptions.Propagation = option
		case "consistent", "cached", "delegated":
			if mount.BindOptions == nil {
				mount.BindOptions = new(types.BindOptions)
			}
			mount.BindOptions.Consistency = option
			// ignore unknown options
		}
	}
	return nil
}

func populateType(mount *types.Mount) {
	switch {
	// Anonymous volume
	case mount.Source == "":
		mount.Type = "volume"
	case isFilePath(mount.Source):
		mount.Type = "bind"
	default:
		mount.Type = "volume"
	}
}

func isFilePath(source string) bool {
	if len(source) == 0 {
		return false
	}
	switch source[0] {
	case '.', '/', '~':
		return true
	}
	// windows named pipes
	if strings.HasPrefix(source, `\\`) {
		return true
	}
	runes := []rune(source)
	if len(runes) < 2 {
		return false
	}
	return isWindowsDrive(runes[:1], runes[1])
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
