package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"

	"drassi.run/core/pkg/container/types"
	"github.com/docker/cli/cli/compose/loader"
	dockermount "github.com/docker/docker/api/types/mount"
	"github.com/docker/go-units"
)

func (fm *flagMapper) mapStorage(copts *containerOptions) error {
	for _, m := range copts.mounts.Value() {
		mount := parseMount(m)
		fm.Spec.Mounts = append(fm.Spec.Mounts, mount)
	}
	for _, v := range copts.volumes.GetAll() {
		if mount, err := ParseVolume(v); err != nil {
			return err
		} else {
			fm.Spec.Mounts = append(fm.Spec.Mounts, mount)
		}
	}
	for _, t := range copts.tmpfs.GetAll() {
		if mount, err := parseTmpfs(t); err != nil {
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

	if storageOpts, err := parseStorageOpts(copts.storageOpt.GetAll()); err != nil {
		return err
	} else {
		fm.Spec.StorageOpt = storageOpts
	}

	fm.Spec.VolumesFrom = copts.volumesFrom.GetAll()
	fm.Spec.ReadonlyRootfs = copts.readonlyRootfs
	return nil
}

func ParseVolume(v string) (*types.Mount, error) {
	parsed, err := loader.ParseVolume(v)
	if err != nil {
		return nil, err
	}
	mount := &types.Mount{
		Type:        parsed.Type,
		Source:      parsed.Source,
		Target:      parsed.Target,
		ReadOnly:    parsed.ReadOnly,
		Consistency: parsed.Consistency,
	}
	if bind := parsed.Bind; bind != nil {
		mount.BindOptions = &types.BindOptions{
			Propagation: bind.Propagation,
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

func parseMount(m dockermount.Mount) *types.Mount {
	mount := &types.Mount{
		Type:        string(m.Type),
		Source:      m.Source,
		Target:      m.Target,
		ReadOnly:    m.ReadOnly,
		Consistency: string(m.Consistency),
	}
	if bo := m.BindOptions; bo != nil {
		mount.BindOptions = &types.BindOptions{
			Propagation:    string(bo.Propagation),
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
	if to := m.TmpfsOptions; to != nil {
		mount.TmpfsOptions = &types.TmpfsOptions{
			Size:    to.SizeBytes,
			Mode:    to.Mode,
			Options: to.Options,
		}
	}
	return mount
}

func parseTmpfs(t string) (*types.Mount, error) {
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

// parses storage options per container into a map
// https://github.com/docker/cli/blob/v27.3.1/cli/command/container/opts.go#L974-L984
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

// ValidateVolumeCtrDir validates a volume mount's destination directory.
func validateVolumeContainerDir(path string) error {
	if path == "" {
		return errors.New("container directory cannot be empty")
	}
	if !filepath.IsAbs(path) {
		return fmt.Errorf("invalid container path %q, must be an absolute path", path)
	}
	return nil
}
