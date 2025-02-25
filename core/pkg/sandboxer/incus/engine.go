package incus

import (
	"context"
	"path"
	"strings"

	c "drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/docker"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/sandboxer/apis/v1alpha1"
	"drassi.run/core/pkg/sandboxer/container"
	"drassi.run/core/util/string"
	dockerclient "github.com/docker/docker/client"
	incusclient "github.com/lxc/incus/v6/client"
	incusapi "github.com/lxc/incus/v6/shared/api"
)

type engine struct {
	client   incusclient.InstanceServer
	template *v1alpha1.IncusTemplate
	source   *incusapi.InstanceSource
}

func New(spec *v1alpha1.IncusSandboxerSpec) (sandboxer.Engine, error) {
	if client, err := incusclient.ConnectIncusUnix(spec.Endpoint, nil); err != nil {
		return nil, err
	} else if source, err := instanceSource(spec.Template.Image); err != nil {
		return nil, err
	} else {
		e := &engine{
			client:   client,
			template: &spec.Template,
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
	name := e.sandboxName(req.Github)
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

func (e *engine) sandboxName(gh *records.Github) string {
	repo := xstring.Normalize(gh.Repository)
	repo = strings.ToLower(repo)

	workflow := strings.TrimSuffix(gh.Workflow, ".yml")
	workflow = strings.TrimSuffix(workflow, ".yaml")
	workflow = xstring.Normalize(workflow)

	job := xstring.Normalize(gh.Job)
	run := xstring.Normalize(gh.RunId)
	attempt := xstring.Normalize(gh.RunAttempt)

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
