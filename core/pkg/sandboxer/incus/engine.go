/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package incus

import (
	"context"
	"io"
	"net"
	"path"
	"strings"
	"sync"

	c "drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/docker"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/sandboxer"
	xsync "drassi.run/core/util/sync"
	incusclient "github.com/lxc/incus/v7/client"
	incusapi "github.com/lxc/incus/v7/shared/api"
	dockerclient "github.com/moby/moby/client"
)

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

type engine struct {
	client      incusclient.InstanceServer
	template    *Template
	source      *incusapi.InstanceSource
	provisioner *provision.Provisioner[*Template]
}

func (e *engine) Launch(ctx context.Context, req *sandboxer.LaunchRequest) (*sandboxer.LaunchResponse, error) {
	tmpl := e.template.Clone()
	tmpl.Name = e.sandboxName(req.Forge)

	launcher := e.launch
	if prov := e.provisioner; prov != nil {
		launcher = prov.Launch(layout.Runtimes(), launcher)
	}
	sb, err := launcher(ctx, tmpl)
	if err != nil {
		return nil, err
	}

	resp := &sandboxer.LaunchResponse{
		Sandbox: sb,
		Mounter: newMounter(),
	}

	var (
		clientMu     sync.Mutex
		clientCloser io.Closer
	)
	if d, ok := sandboxer.Unwrap(sb).(interface {
		Dialer(cmd []string) func(ctx context.Context, network, addr string) (net.Conn, error)
	}); ok {
		resp.Sandbox = sandboxer.AddBeforeCleanup(resp.Sandbox, func(context.Context) error {
			clientMu.Lock()
			defer clientMu.Unlock()
			if clientCloser != nil {
				return clientCloser.Close()
			}
			return nil
		})

		resp.ContainerEngine = xsync.Singleton(func(context.Context) (c.Engine, error) {
			dialer := d.Dialer(docker.ProxyCommand(""))
			client, err := docker.New(dockerclient.WithDialContext(dialer))
			if err != nil {
				return nil, err
			}
			clientMu.Lock()
			clientCloser = client
			clientMu.Unlock()
			return c.WithTelemetry(client), nil
		})
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
	name := forge.CanonicalName()
	name = strings.ReplaceAll(name, "_", "-")
	return name
}

func (e *engine) Close() error {
	e.client.Disconnect()
	return nil
}
