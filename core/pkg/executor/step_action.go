/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package executor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"

	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/actions"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/pkg/store/repository"
	"drassi.run/core/pkg/store/repository/gitstore"
	"drassi.run/core/util/dig"
	"drassi.run/core/util/otel"
	"github.com/go-git/go-git/v5/plumbing/object"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/dig"
	"gopkg.in/yaml.v3"
)

type ActionStepDef struct {
	// Example:
	// + Using a public action: `uses: actions/aws@v2.0.1`
	// + Using a public action in a subdirectory: `uses: actions/aws/ec2@main`
	// + Using a local action: `uses: ./.github/actions/hello-world-action`
	Repo *repository.Repository

	// cache resolved revision
	rev string
}

func (d *ActionStepDef) Repository() *repository.Repository {
	return d.Repo
}

func (d *ActionStepDef) PrepareExecute(ctx context.Context, scope *dig.Scope) (StepRun, error) {
	s := scribe.FromContext(ctx)

	var (
		github  records.Github
		store   gitstore.Store
		sandbox sandboxer.Sandbox
		exprEnv expression.Env
	)

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(xotel.ActionRepo(repository.Location(d.Repo)))

	if err := xdig.Populate(scope, &github); err != nil {
		return nil, err
	}
	if err := xdig.Populate(scope, &store); err != nil {
		return nil, err
	}
	if err := xdig.Populate(scope, &sandbox); err != nil {
		return nil, err
	}
	if err := xdig.Populate(scope, &exprEnv); err != nil {
		return nil, err
	}
	if err := xdig.Supply(scope, d.Repo); err != nil {
		return nil, err
	}

	token := github.Token
	// If the action is located in different server than the job repo,
	// unset the token to prevent an unauthenticated error.
	if repository.Endpoint(d.Repo) != d.serverDomain(github.ServerUrl) {
		token = ""
	}

	// evaluating before loadAction so it DisplayName can be push-down into child StepRun
	//if err := sr.evaluateDisplayName(ctx, exprEnv, repository.Location(sr.Repo)); err != nil {
	//	return err
	//}

	if rev, err := store.Fetch(ctx, d.Repo, token); err != nil {
		return nil, err
	} else {
		s.Writef("Download action repository %q (SHA:%s)", repository.Location(d.Repo), rev)
		d.rev = rev
	}

	if def, err := d.loadAction(ctx, s, store); err != nil {
		return nil, err
	} else if err = d.transferAction(ctx, store, sandbox); err != nil {
		return nil, err
	} else {
		return def.PrepareExecute(ctx, scope)
	}
}

func (d *ActionStepDef) loadAction(ctx context.Context, s *scribe.Scribe, store gitstore.Store) (StepDef, error) {
	span := trace.SpanFromContext(ctx)

	// 1. First, try reading "action.yml" or "action.yaml" file
	for _, f := range []string{"action.yml", "action.yaml"} {
		path := filepath.Join(d.Repo.Path, f)
		if r, err := store.File(ctx, d.Repo, d.rev, path); err == nil {
			span.AddEvent("Loaded Action",
				trace.WithAttributes(xotel.ActionPath(path)),
			)
			s.Debugf("Loading %q for action", path)
			return d.loadActionManifest(r)
		} else if !errors.Is(err, object.ErrFileNotFound) {
			return nil, err
		}
	}

	// 2. Second, try reading "Dockerfile" or "dockerfile"
	for _, f := range []string{"Dockerfile", "dockerfile"} {
		path := filepath.Join(d.Repo.Path, f)
		if r, err := store.File(ctx, d.Repo, d.rev, path); err == nil {
			r.Close()
			span.AddEvent("Loaded Action",
				trace.WithAttributes(xotel.ActionPath(path)),
			)
			s.Debugf("Loading %q for action", path)
			return d.createDockerfileAction(path)
		} else if !errors.Is(err, object.ErrFileNotFound) {
			return nil, err
		}
	}

	return nil, fmt.Errorf(`file "action.yml", "action.yaml" and "Dockerfile" not found in your given path`)
}

func (d *ActionStepDef) loadActionManifest(r io.ReadCloser) (StepDef, error) {
	defer r.Close()

	m := make(map[string]any)
	if err := yaml.NewDecoder(r).Decode(m); err != nil {
		return nil, err
	}
	action := new(actions.Action)
	if err := model.Decode(m, action); err != nil {
		return nil, err
	}

	return ToStepDef(action)
}

func (d *ActionStepDef) createDockerfileAction(dockerfile string) (StepDef, error) {
	def := &DockerStepDef{
		Image: dockerfile,
	}
	return def, nil
}

func (d *ActionStepDef) transferAction(ctx context.Context, store gitstore.Store, sandbox sandboxer.Sandbox) error {
	location := repository.FullName(d.Repo) + "@" + d.Repo.Ref
	r, err := store.Read(ctx, d.Repo, d.rev, location)
	if err != nil {
		return err
	}
	defer r.Close()

	return sandbox.CopyIn(ctx, r, sandbox.Layout().Actions)
}

func (d *ActionStepDef) serverDomain(s string) string {
	if u, err := url.Parse(s); err == nil {
		return u.Host
	}
	return ""
}
