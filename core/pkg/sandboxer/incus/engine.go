package incus

import (
	"context"
	"strings"

	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/util/string"
	incusclient "github.com/lxc/incus/v6/client"
	incusapi "github.com/lxc/incus/v6/shared/api"
)

type engine struct {
	client   incusclient.InstanceServer
	template *IncusTemplate
}

func New(spec *IncusSpec) (sandboxer.Engine, error) {
	client, err := incusclient.ConnectIncusUnix(spec.Endpoint, nil)
	if err != nil {
		return nil, err
	}
	e := &engine{
		client:   client,
		template: &spec.Template,
	}
	return e, nil
}

func (e *engine) Close() error {
	e.client.Disconnect()
	return nil
}

func (e *engine) Launch(ctx context.Context, req *sandboxer.LaunchRequest) (*sandboxer.LaunchResponse, error) {
	name := e.sandboxName(req)
	iReq := incusapi.InstancesPost{
		Name:         name,
		Start:        true,
		Source:       e.template.Source,
		Type:         e.template.Type,
		InstanceType: e.template.InstanceSize,
		InstancePut: incusapi.InstancePut{
			Architecture: e.template.Architecture,
			Config:       e.template.Config,
			Devices:      e.template.Devices,
			Ephemeral:    e.template.Ephemeral,
			Profiles:     e.template.Profiles,
			Restore:      e.template.Restore,
			Stateful:     e.template.Stateful,
			Description:  e.template.Description,
		},
	}
	if op, err := e.client.CreateInstance(iReq); err != nil {
		return nil, err
	} else if err = op.WaitContext(ctx); err != nil {
		return nil, err
	}

	if sb, err := newSandbox(e.client, name); err != nil {
		return nil, err
	} else {
		resp := &sandboxer.LaunchResponse{
			Sandbox: sb,
		}
		return resp, nil
	}
}

func (e *engine) sandboxName(req *sandboxer.LaunchRequest) string {
	repo := xstring.Normalize(req.Github.Repository)
	repo = strings.ToLower(repo)

	workflow := strings.TrimSuffix(req.Github.Workflow, ".yml")
	workflow = strings.TrimSuffix(workflow, ".yaml")
	workflow = xstring.Normalize(workflow)

	job := xstring.Normalize(req.Github.Job)
	run := xstring.Normalize(req.Github.RunId)
	attempt := xstring.Normalize(req.Github.RunAttempt)

	name := strings.Join([]string{repo, workflow, job, run, attempt}, "-")
	name = strings.ReplaceAll(name, "_", "-")
	return name
}
