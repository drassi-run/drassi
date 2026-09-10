/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package incus

import (
	"context"
	"errors"
	"maps"
	"net"
	"path"
	"slices"
	"strings"
	"sync"

	"drassi.run/core/config"
	c "drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/docker"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/sandboxer/container"
	"drassi.run/core/pkg/store/oci"
	"drassi.run/core/util/string"
	incusclient "github.com/lxc/incus/v6/client"
	incusapi "github.com/lxc/incus/v6/shared/api"
	dockerclient "github.com/moby/moby/client"
)

func init() {
	sandboxer.Register(config.ProviderIncus, DefaultConfig, NewFactory)
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

func NewFactory(cfg *Config) sandboxer.Factory {
	f := &factory{cfg: cfg}
	f.create = sync.OnceValues(f.doCreate)
	return f
}

type factory struct {
	create func() (sandboxer.Engine, error)

	cfg      *Config
	store    ocistore.Manager
	runtimes map[string]*config.Runtime
}

func (f *factory) SetOciStore(store ocistore.Manager) {
	f.store = store
}

func (f *factory) ProvisionRuntime(config map[string]*config.Runtime) {
	f.runtimes = config
}

func (f *factory) Create() (sandboxer.Engine, error) {
	return f.create()
}

func (f *factory) doCreate() (sandboxer.Engine, error) {
	var prov *provision.Provisioner[*Template]
	if len(f.runtimes) > 0 {
		if f.store == nil {
			return nil, errors.New("oci store is required when runtimes are configured")
		}
		prov = provision.New[*Template](
			f.runtimes,
			provision.Pull[*Template](f.store),
			provision.Mount[*Template](f.store, ocistore.WithWritable(true)),
			AddDiskDevice(),
		)
	}
	return New(f.cfg, prov)
}

type Config struct {
	Endpoint string   `toml:"endpoint" json:"endpoint"`
	Template Template `toml:"template" json:"template"`
}

// Template for create incus VM
// [github.com/lxc/incus/v6/shared/api.InstancesPost]
type Template struct {
	Name string `toml:"name,omitempty" json:"name,omitempty"`

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

func (t *Template) Clone() *Template {
	if t == nil {
		return nil
	}
	c := *t
	if t.Profiles != nil {
		c.Profiles = slices.Clone(t.Profiles)
	}
	if t.Config != nil {
		c.Config = maps.Clone(t.Config)
	}
	if t.Devices != nil {
		c.Devices = make(map[string]map[string]string, len(t.Devices))
		for k, v := range t.Devices {
			c.Devices[k] = maps.Clone(v)
		}
	}
	return &c
}

type engine struct {
	client      incusclient.InstanceServer
	template    *Template
	source      *incusapi.InstanceSource
	provisioner *provision.Provisioner[*Template]
}

func New(config *Config, prov *provision.Provisioner[*Template]) (sandboxer.Engine, error) {
	if client, err := incusclient.ConnectIncusUnix(config.Endpoint, nil); err != nil {
		return nil, err
	} else if source, err := instanceSource(config.Template.Image); err != nil {
		return nil, err
	} else {
		e := &engine{
			client:      client,
			template:    &config.Template,
			source:      source,
			provisioner: prov,
		}
		return e, nil
	}
}

func (e *engine) Launch(ctx context.Context, req *sandboxer.LaunchRequest) (*sandboxer.LaunchResponse, error) {
	tmpl := e.template.Clone()
	tmpl.Name = e.sandboxName(req.Forge)

	launcher := e.launch
	if prov := e.provisioner; prov != nil {
		launcher = prov.Launch(defaultRuntimeDir, launcher)
	}
	sb, err := launcher(ctx, tmpl)
	if err != nil {
		return nil, err
	}

	d, ok := sandboxer.Unwrap(sb).(interface {
		Dialer(cmd []string) func(ctx context.Context, network, addr string) (net.Conn, error)
	})
	if !ok {
		_ = sb.Terminate(ctx)
		return nil, errors.New("sandbox does not support dialer")
	}
	dialer := d.Dialer(docker.ProxyCommand(""))
	client, err := docker.New(dockerclient.WithDialContext(dialer))
	if err != nil {
		_ = sb.Terminate(ctx)
		return nil, err
	}
	client = c.WithTelemetry(client)

	s := sandboxer.AddBeforeCleanup(sb, func(context.Context) error {
		return client.Close()
	})
	b := container.NewBootstrapper(client)
	resp, err := b.Bootstrap(ctx, s, req)
	if err != nil {
		_ = s.Terminate(ctx)
		return nil, err
	}
	return resp, nil
}

func (e *engine) launch(ctx context.Context, tmpl *Template) (sandboxer.Sandbox, error) {
	iReq := incusapi.InstancesPost{
		Name:         tmpl.Name,
		Start:        true,
		Source:       *e.source,
		Type:         incusapi.InstanceTypeContainer,
		InstanceType: tmpl.InstanceSize,
		InstancePut: incusapi.InstancePut{
			Architecture: tmpl.Architecture,
			Config:       tmpl.Config,
			Devices:      tmpl.Devices,
			Ephemeral:    tmpl.Ephemeral,
			Profiles:     tmpl.Profiles,
		},
	}
	if op, err := e.client.CreateInstance(iReq); err != nil {
		return nil, err
	} else if err = op.WaitContext(ctx); err != nil {
		return nil, err
	}

	sb, err := newSandbox(e.client, tmpl.Name)
	if err != nil {
		return nil, err
	}
	return sb, nil
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

func (e *engine) Close() error {
	e.client.Disconnect()
	return nil
}
