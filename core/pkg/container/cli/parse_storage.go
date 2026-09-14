/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"fmt"
	"strings"

	"drassi.run/core/pkg/container/parser"
	"drassi.run/core/pkg/container/types"
	dockermount "github.com/moby/moby/api/types/mount"
)

func (fm *flagMapper) mapStorage(copts *containerOptions) error {
	for _, m := range copts.mounts.Value() {
		mount := parseMount(m)
		fm.Spec.Mounts = append(fm.Spec.Mounts, mount)
	}
	for _, v := range copts.volumes.GetAllOrEmpty() {
		if mount, err := parser.ParseVolume(v); err != nil {
			return err
		} else {
			fm.Spec.Mounts = append(fm.Spec.Mounts, mount)
		}
	}
	for _, t := range copts.tmpfs.GetAllOrEmpty() {
		if mount, err := parser.ParseTmpfs(t); err != nil {
			return err
		} else {
			fm.Spec.Mounts = append(fm.Spec.Mounts, mount)
		}
	}

	if driver := copts.volumeDriver; driver != "" {
		for _, m := range fm.Spec.Mounts {
			if m.Type != "volume" {
				continue
			}
			if m.VolumeOptions == nil {
				m.VolumeOptions = new(types.VolumeOptions)
			}
			m.VolumeOptions.Driver = driver
		}
	}

	if storageOpts, err := parseStorageOpts(copts.storageOpt.GetAllOrEmpty()); err != nil {
		return err
	} else {
		fm.Spec.StorageOpt = storageOpts
	}

	fm.Spec.VolumesFrom = copts.volumesFrom.GetAllOrEmpty()
	fm.Spec.ReadonlyRootfs = copts.readonlyRootfs
	return nil
}

func parseMount(m dockermount.Mount) *types.Mount {
	mount := &types.Mount{
		Type:     string(m.Type),
		Source:   m.Source,
		Target:   m.Target,
		ReadOnly: m.ReadOnly,
	}
	if bo := m.BindOptions; bo != nil {
		mount.BindOptions = &types.BindOptions{
			Propagation:    string(bo.Propagation),
			Consistency:    string(m.Consistency),
			CreateHostPath: bo.CreateMountpoint,
		}
	}
	if vo := m.VolumeOptions; vo != nil {
		mount.VolumeOptions = &types.VolumeOptions{
			NoCopy:  vo.NoCopy,
			SubPath: vo.Subpath,
		}
		if dc := vo.DriverConfig; dc != nil {
			mount.VolumeOptions.Driver = dc.Name
			mount.VolumeOptions.Options = dc.Options
		}
	}
	if io := m.ImageOptions; io != nil {
		mount.ImageOptions = &types.ImageOptions{
			Subpath: io.Subpath,
		}
	}
	if to := m.TmpfsOptions; to != nil {
		mount.TmpfsOptions = &types.TmpfsOptions{
			Size:    to.SizeBytes,
			Mode:    to.Mode,
			Options: to.Options,
		}
	}
	return mount
}

// parses storage options per container into a map
// https://github.com/docker/cli/blob/v29.7.2/cli/command/container/opts.go#L981-L992
func parseStorageOpts(storageOpts []string) (map[string]string, error) {
	m := make(map[string]string)
	for _, option := range storageOpts {
		k, v, ok := strings.Cut(option, "=")
		if !ok {
			return nil, fmt.Errorf("invalid storage option")
		}
		m[k] = v
	}
	return m, nil
}
