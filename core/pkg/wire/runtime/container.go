package wire_runtime

import (
	"context"
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"
	"time"

	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/executor/runtime"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/util/string"
	. "drassi.run/core/util/types"
)

func NewContainerRuntime(
	ctx context.Context,
	engine container.Engine,
	streams container.Streams,
	sandbox sandboxer.Sandbox,
	info *records.JobInfo,
	gh *records.Github,
) (runtime.Container, error) {
	opts := make([]runtime.ContainerRuntimeOption, 0)
	if opt := labelsOpt(gh); opt != nil {
		opts = append(opts, opt)
	}
	if opt := networkOpt(info); opt != nil {
		opts = append(opts, opt)
	}

	layout := sandbox.Layout()
	if ctn := info.Container; ctn != nil {
		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		var mounts []*types.Mount
		if spec, err := engine.ContainerInspect(ctx, ctn.Id); err != nil {
			return nil, err
		} else {
			mounts = spec.Mounts
		}

		if opt, err := containerMountOpt(layout, mounts); err != nil {
			return nil, err
		} else {
			opts = append(opts, opt)
		}
	} else {
		opt := sandboxMountOpt(layout)
		opts = append(opts, opt)
	}

	return runtime.NewContainerRuntime(engine, streams, opts...)
}

func labelsOpt(gh *records.Github) runtime.ContainerRuntimeOption {
	labels := types.LabelsFor(gh)
	return runtime.WithLabels(labels)
}

func networkOpt(info *records.JobInfo) runtime.ContainerRuntimeOption {
	if info.Container != nil {
		if net := info.Container.Network; net != "" {
			return runtime.WithNetwork(net)
		}
	}

	for _, svc := range info.Services {
		if net := svc.Network; net != "" {
			return runtime.WithNetwork(net)
		}
	}

	return nil
}

func sandboxMountOpt(layout *sandboxer.Layout) runtime.ContainerRuntimeOption {
	wsMount := &types.Mount{
		Type:   "bind",
		Source: layout.Workspace,
		Target: "/opt/drassi/workspace",
	}
	tmpMount := &types.Mount{
		Type:   "bind",
		Source: layout.Temp,
		Target: "/opt/drassi/temp",
	}

	mounts := []Pair[string, *types.Mount]{
		{Key: layout.Workspace, Value: wsMount},
		{Key: layout.Temp, Value: tmpMount},
	}
	return runtime.WithMounts(mounts)
}

func containerMountOpt(layout *sandboxer.Layout, sbMounts []*types.Mount) (runtime.ContainerRuntimeOption, error) {
	slices.SortFunc(sbMounts, func(a, b *types.Mount) int {
		return strings.Compare(a.Target, b.Target)
	})

	mounts := make([]Pair[string, *types.Mount], 0)
	if mount, subDir := mountOf(layout.Workspace, sbMounts); mount == nil {
		return nil, fmt.Errorf("workspace dir %s is not in a mount point", layout.Workspace)
	} else if m, err := newMount(mount, "/opt/drassi/workspace", subDir); err != nil {
		return nil, err
	} else {
		mounts = append(mounts, Pair[string, *types.Mount]{
			Key:   mount.Target, // sandboxPath
			Value: m,            // containerMount
		})
	}

	if mount, subDir := mountOf(layout.Temp, sbMounts); mount == nil {
		return nil, fmt.Errorf("temp dir %s is not in a mount point", layout.Workspace)
	} else if m, err := newMount(mount, "/opt/drassi/temp", subDir); err != nil {
		return nil, err
	} else {
		mounts = append(mounts, Pair[string, *types.Mount]{
			Key:   mount.Target, // sandboxPath
			Value: m,            // containerMount
		})
	}

	return runtime.WithMounts(mounts), nil
}

func newMount(m *types.Mount, point, subDir string) (*types.Mount, error) {
	mount := &types.Mount{
		Type:     m.Type,
		Source:   m.Source,
		Target:   point,
		ReadOnly: m.ReadOnly,
	}

	switch typ := m.Type; typ {
	case "bind":
		if bind := m.BindOptions; bind != nil {
			mount.BindOptions = &types.BindOptions{
				Propagation:    bind.Propagation,
				Consistency:    bind.Consistency,
				Recursive:      bind.Recursive,
				CreateHostPath: bind.CreateHostPath,
			}
		}
		if subDir != "" {
			mount.Source = path.Join(m.Source, subDir)
		}
	case "volume":
		if vol := m.VolumeOptions; vol != nil {
			mount.VolumeOptions = &types.VolumeOptions{
				NoCopy:  vol.NoCopy,
				Labels:  maps.Clone(vol.Labels),
				SubPath: vol.SubPath,
				Driver:  vol.Driver,
				Options: maps.Clone(vol.Options),
			}
			if subDir != "" {
				mount.VolumeOptions.SubPath = path.Join(mount.VolumeOptions.SubPath, subDir)
			}
		} else if subDir != "" {
			mount.VolumeOptions = &types.VolumeOptions{
				SubPath: subDir,
			}
		}
	default:
		return nil, fmt.Errorf("unsupported mount type: %s", typ)
	}

	return mount, nil
}

func mountOf(path string, mounts []*types.Mount) (*types.Mount, string) {
	strippedOrigin := strings.TrimRight(path, "/")
	for _, m := range mounts {
		target := m.Target
		if strings.TrimRight(target, "/") == strippedOrigin {
			return m, ""
		}

		target = xstring.EnsureSuffix(target, "/")
		if inner, ok := strings.CutPrefix(path, target); ok {
			return m, inner
		}
	}

	return nil, ""
}
