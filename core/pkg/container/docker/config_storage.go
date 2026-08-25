/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package docker

import (
	"drassi.run/core/pkg/container/cli"
	"drassi.run/core/pkg/container/types"
	dockercontainer "github.com/moby/moby/api/types/container"
	dockermount "github.com/moby/moby/api/types/mount"
)

func (cc *containerConfig) setStorage(conf *types.ContainerStorage) {
	hc := cc.HostConfig
	hc.ReadonlyRootfs = conf.ReadonlyRootfs
	hc.StorageOpt = conf.StorageOpt
	hc.VolumesFrom = conf.VolumesFrom
	cc.setMounts(conf.Mounts)
}

func (cc *containerConfig) setMounts(volumes []*types.Mount) {
	mounts := make([]dockermount.Mount, len(volumes))
	for i, v := range volumes {
		m := dockermount.Mount{
			Type:     dockermount.Type(v.Type),
			Source:   v.Source,
			Target:   v.Target,
			ReadOnly: v.ReadOnly,
		}
		if bind := v.BindOptions; bind != nil {
			m.Consistency = dockermount.Consistency(bind.Consistency)
			m.BindOptions = &dockermount.BindOptions{
				Propagation:      dockermount.Propagation(bind.Propagation),
				CreateMountpoint: bind.CreateHostPath,
			}
		}
		if volume := v.VolumeOptions; volume != nil {
			m.VolumeOptions = &dockermount.VolumeOptions{
				NoCopy:  volume.NoCopy,
				Subpath: volume.SubPath,
			}
		}
		if tmpfs := v.TmpfsOptions; tmpfs != nil {
			m.TmpfsOptions = &dockermount.TmpfsOptions{
				SizeBytes: tmpfs.Size,
				Mode:      tmpfs.Mode,
			}
		}
		mounts[i] = m
	}
	if len(mounts) > 0 {
		cc.HostConfig.Mounts = mounts
	}
	// anonymous volumes, e.g.
	// --volume /path/in/container
	// --mount type=volume,destination=/path/in/container
	cc.Config.Volumes = nil
	// --volume named-volume:/path/in/container
	// --volume /path/on/host:/path/in/container
	// --mount type=volume,source=my-volume,destination=/path/in/container
	// --mount type=bind,source=/path/on/host,destination=/path/in/container
	cc.HostConfig.Binds = nil
	cc.HostConfig.VolumeDriver = ""
	// --tmpfs /tmp:rw,size=787448k,mode=1777
	// --mount type=tmpfs,destination=/path/in/container
	cc.HostConfig.Tmpfs = nil
}

func (cs *containerSpec) setStorage(info dockercontainer.InspectResponse) error {
	hc := info.HostConfig

	if err := cs.setTmpfs(hc.Tmpfs); err != nil {
		return err
	}
	if err := cs.setVolumes(hc.Binds); err != nil {
		return err
	}
	cs.setMounts(hc.Mounts)
	cs.setResolvedMounts(info.Mounts)

	return nil
}

func (cs *containerSpec) setTmpfs(tmpfs map[string]string) error {
	for k, v := range tmpfs {
		if mount, err := cli.ParseTmpfs(k + ":" + v); err != nil {
			return err
		} else {
			cs.Spec.Mounts = append(cs.Spec.Mounts, mount)
		}
	}
	return nil
}

func (cs *containerSpec) setVolumes(volumes []string) error {
	for _, v := range volumes {
		if mount, err := cli.ParseVolume(v); err != nil {
			return err
		} else {
			cs.Spec.Mounts = append(cs.Spec.Mounts, mount)
		}
	}
	return nil
}

func (cs *containerSpec) setMounts(mounts []dockermount.Mount) {
	for _, m := range mounts {
		mount := &types.Mount{
			Type:     string(m.Type),
			Source:   m.Source,
			Target:   m.Target,
			ReadOnly: m.ReadOnly,
		}

		switch m.Type {
		case dockermount.TypeBind:
			if bind := m.BindOptions; bind != nil {
				mount.BindOptions = &types.BindOptions{
					Propagation:    string(bind.Propagation),
					CreateHostPath: bind.CreateMountpoint,
				}
				// TODO
				//opts := mount.BindOptions
				//if bind.NonRecursive {
				//	opts.Recursive = "disabled"
				//}
				//if bind.ReadOnlyNonRecursive {
				//	opts.Recursive = "readonly"
				//}
			}
			if m.Consistency != "" && m.Consistency != dockermount.ConsistencyDefault {
				if mount.BindOptions == nil {
					mount.BindOptions = new(types.BindOptions)
				}
				mount.BindOptions.Consistency = string(m.Consistency)
			}
		case dockermount.TypeVolume:
			if volume := m.VolumeOptions; volume != nil {
				mount.VolumeOptions = &types.VolumeOptions{
					NoCopy:  volume.NoCopy,
					Labels:  volume.Labels,
					SubPath: volume.Subpath,
				}
				if driver := volume.DriverConfig; driver != nil {
					opts := mount.VolumeOptions
					opts.Driver = driver.Name
					opts.Options = driver.Options
				}
			}
		case dockermount.TypeTmpfs:
			if tmpfs := m.TmpfsOptions; tmpfs != nil {
				mount.TmpfsOptions = &types.TmpfsOptions{
					Size:    tmpfs.SizeBytes,
					Mode:    tmpfs.Mode,
					Options: tmpfs.Options,
				}
			}
		}

		cs.Spec.Mounts = append(cs.Spec.Mounts, mount)
	}
}

// * add Driver info into volume mount
// * add anonymous volume
// * volumesFrom
func (cs *containerSpec) setResolvedMounts(mounts []dockercontainer.MountPoint) {
	mountMap := make(map[string]*types.Mount)
	for _, m := range cs.Spec.Mounts {
		// mount targets are uniq in a container
		mountMap[m.Target] = m
	}

	for _, m := range mounts {
		if mount, ok := mountMap[m.Destination]; ok {
			if m.Type == dockermount.TypeVolume && m.Driver != "" {
				if mount.VolumeOptions == nil {
					mount.VolumeOptions = new(types.VolumeOptions)
				}
				mount.VolumeOptions.Driver = m.Driver
			}
			continue
		}

		mount := &types.Mount{
			Type:     string(m.Type),
			Target:   m.Destination,
			ReadOnly: !m.RW,
		}
		switch m.Type {
		case dockermount.TypeVolume:
			mount.Source = m.Name
			if driver := m.Driver; driver != "" {
				mount.VolumeOptions = &types.VolumeOptions{
					Driver: driver,
				}
			}
		case dockermount.TypeBind:
			if m.Propagation != "" {
				mount.BindOptions = &types.BindOptions{
					Propagation: string(m.Propagation),
				}
			}
			fallthrough
		default:
			mount.Source = m.Source
			// TODO: parse Mode
		}
	}
}
