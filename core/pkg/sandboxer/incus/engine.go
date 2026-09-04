/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package incus

import (
	"context"
	"path"
	"strings"

	c "drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/docker"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/sandboxer/container"
	"drassi.run/core/util/string"
	incusclient "github.com/lxc/incus/v6/client"
	incusapi "github.com/lxc/incus/v6/shared/api"
	dockerclient "github.com/moby/moby/client"
)

type Config struct {
	Endpoint string   `toml:"endpoint" json:"endpoint"`
	Template Template `toml:"template" json:"template"`
}

// Template for create incus VM
// [github.com/lxc/incus/v6/shared/api.InstancesPost]
type Template struct {
	// OCI image name, e.g: ghcr.io/drassi-run/ubuntu:22.04
	Image string `toml:"source" json:"image"`

	// Instance architecture, e.g: x86_64
	Architecture string `toml:"architecture" json:"architecture,omitempty"`

	// Cloud instance size (AWS, GCP, Azure, ...) to emulate with limits
	// Example: t1.micro
	InstanceSize string `toml:"instance_size" json:"instance_size,omitempty"`

	// List of profiles applied to the instance
	// Example: ["default"]
	Profiles []string `toml:"profiles" json:"profiles,omitempty"`

	// Instance configuration (see https://linuxcontainers.org/incus/docs/main/instances/)
	// Example: {"security.nesting": "true"}
	Config map[string]string `toml:"config" json:"config,omitempty"`

	// Instance devices (see https://linuxcontainers.org/incus/docs/main/instances/)
	// Example: {"root": {"type": "disk", "pool": "default", "path": "/"}}
	Devices map[string]map[string]string `toml:"devices" json:"devices,omitempty"`

	// Whether the instance is ephemeral (deleted on shutdown)
	// Example: false
	Ephemeral bool `toml:"ephemeral" json:"ephemeral,omitempty"`
}

func DefaultConfig() *Config {
	return &Config{
		Endpoint: "unix:///var/lib/incus/unix.socket",
		Template: Template{
			Image:     "ubuntu:latest",
			Ephemeral: true,
		},
	}
}

type engine struct {
	client   incusclient.InstanceServer
	template *Template
	source   *incusapi.InstanceSource
}

func New(config *Config) (sandboxer.Engine, error) {
	if client, err := incusclient.ConnectIncusUnix(config.Endpoint, nil); err != nil {
		return nil, err
	} else if source, err := instanceSource(config.Template.Image); err != nil {
		return nil, err
	} else {
		e := &engine{
			client:   client,
			template: &config.Template,
			source:   source,
		}
		return e, nil
	}
}

func (e *engine) Close() error {
	e.client.Disconnect()
	return nil
}

func (e *engine) Launch(ctx context.Context, req *sandboxer.LaunchRequest) (*sandboxer.LaunchResponse, error) {
	name := e.sandboxName(req.Forge)
	iReq := incusapi.InstancesPost{
		Name:         name,
		Start:        true,
		Source:       *e.source,
		Type:         incusapi.InstanceTypeContainer,
		InstanceType: e.template.InstanceSize,
		InstancePut: incusapi.InstancePut{
			Architecture: e.template.Architecture,
			Config:       e.template.Config,
			Devices:      e.template.Devices,
			Ephemeral:    e.template.Ephemeral,
			Profiles:     e.template.Profiles,
		},
	}
	if op, err := e.client.CreateInstance(iReq); err != nil {
		return nil, err
	} else if err = op.WaitContext(ctx); err != nil {
		return nil, err
	}

	sb, err := newSandbox(e.client, name)
	if err != nil {
		return nil, err
	}

	dialer := sb.Dialer(docker.ProxyCommand(""))
	client, err := docker.New(dockerclient.WithDialContext(dialer))
	if err != nil {
		return nil, err
	}
	client = c.WithTelemetry(client)

	s := sandboxer.AddBeforeCleanup(sb, func(context.Context) error {
		return client.Close()
	})
	b := container.NewBootstrapper(client)
	return b.Bootstrap(ctx, s, req)
}

func (e *engine) sandboxName(forge *records.Forge) string {
	repo := xstring.Normalize(forge.Repository)
	repo = strings.ToLower(repo)

	workflow := strings.TrimSuffix(forge.Workflow, ".yml")
	workflow = strings.TrimSuffix(workflow, ".yaml")
	workflow = xstring.Normalize(workflow)

	job := xstring.Normalize(forge.Job)
	run := xstring.Normalize(forge.RunId)
	attempt := xstring.Normalize(forge.RunAttempt)

	name := strings.Join([]string{repo, workflow, job, run, attempt}, "-")
	name = strings.ReplaceAll(name, "_", "-")
	return name
}

func instanceSource(uri string) (*incusapi.InstanceSource, error) {
	registry, image, found := strings.Cut(uri, "/")
	if !found || !strings.Contains(registry, ".") {
		registry = "https://docker.io"
		if found {
			image = uri
		} else {
			image = path.Join("library", uri)
		}
	}
	source := &incusapi.InstanceSource{
		Type:     "image",
		Alias:    image,
		Server:   registry,
		Protocol: "oci",
	}
	return source, nil
}
