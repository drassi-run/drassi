package container

import (
	"context"
	"maps"
	"net/url"
	"strings"

	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/sandboxer"
)

const (
	LabelRepository = "run.drassi.repository"
	LabelReference  = "run.drassi.reference"
	LabelWorkflow   = "run.drassi.workflow"
	LabelJob        = "run.drassi.job"
	LabelAttempt    = "run.drassi.attempt"
	LabelRun        = "run.drassi.run"
)

func LabelsFor(gh *records.Github) map[string]string {
	repo := gh.Repository
	if u, err := url.Parse(gh.ServerUrl); err == nil {
		if server := u.Host; server != "" {
			server = strings.ToLower(server)
			server = strings.TrimRight(server, "/")
			repo = server + "/" + repo
		}
	}

	labels := map[string]string{
		LabelRepository: repo,          // e.g: github.com/drassi-run/drassi
		LabelReference:  gh.Ref,        // e.g: refs/heads/main
		LabelWorkflow:   gh.Workflow,   // e.g: test
		LabelJob:        gh.Job,        // e.g: unittests
		LabelAttempt:    gh.RunAttempt, // e.g: 1
		LabelRun:        gh.RunId,      // e.g: 11208400917
	}
	return labels
}

func cleanup(labels map[string]string, fn func(context.Context, *container.RemoveOptions) error) sandboxer.Cleanup {
	return func(ctx context.Context) error {
		return fn(ctx, &container.RemoveOptions{Labels: labels})
	}
}

type refiner = func(*types.ContainerSpec)

func setLabels(labels map[string]string) refiner {
	return func(spec *types.ContainerSpec) {
		// set labels for container
		if spec.Labels == nil {
			spec.Labels = maps.Clone(labels)
		} else {
			maps.Copy(spec.Labels, labels)
		}

		// set labels for volumes
		for _, vol := range spec.Mounts {
			if vol.Type != "volume" {
				continue
			}
			if vol.VolumeOptions == nil {
				vol.VolumeOptions = &types.VolumeOptions{}
			}
			if opts := vol.VolumeOptions; opts.Labels == nil {
				opts.Labels = maps.Clone(labels)
			} else {
				maps.Copy(opts.Labels, labels)
			}
		}
	}
}

func setNetwork(id string) refiner {
	return func(spec *types.ContainerSpec) {
	}
}

func setCmd(entrypoint, command []string) refiner {
	return func(spec *types.ContainerSpec) {
		if entrypoint != nil {
			spec.Entrypoint = entrypoint
		}
		if command != nil {
			spec.Command = command
		}
	}
}
