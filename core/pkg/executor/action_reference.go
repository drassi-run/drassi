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

type ReferenceActionSpec struct {
	// Example:
	// + Using a public action: `uses: actions/aws@v2.0.1`
	// + Using a public action in a subdirectory: `uses: actions/aws/ec2@main`
	// + Using a local action: `uses: ./.github/actions/hello-world-action`
	Repo *repository.Repository

	// cache resolved revision
	rev string
}

func (spec *ReferenceActionSpec) Repository() *repository.Repository {
	return spec.Repo
}

func (spec *ReferenceActionSpec) CreateExecutor(
	ctx context.Context, scope *dig.Scope, exec StepExecutor,
) (ActionExecutor, error) {
	s := scribe.FromContext(ctx)

	var (
		github *records.Github
		store  gitstore.Store
	)

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(xotel.ActionRepo(repository.Location(spec.Repo)))

	if err := xdig.Populate(scope, &github); err != nil {
		return nil, err
	}
	if err := xdig.Populate(scope, &store); err != nil {
		return nil, err
	}
	if err := xdig.Supply(scope, spec.Repo); err != nil {
		return nil, err
	}

	token := github.Token
	// If the action is located in different server than the job repo,
	// unset the token to prevent an unauthenticated error.
	if repository.Endpoint(spec.Repo) != spec.serverDomain(github.ServerUrl) {
		token = ""
	}

	if rev, err := store.Fetch(ctx, spec.Repo, token); err != nil {
		return nil, err
	} else {
		s.Writef("Download action repository %q (SHA:%s)", repository.Location(spec.Repo), rev)
		spec.rev = rev
	}

	if action, err := spec.loadAction(ctx, s, store); err != nil {
		return nil, err
	} else if err = spec.transferAction(ctx, store, exec.Sandbox()); err != nil {
		return nil, err
	} else {
		return action.CreateExecutor(ctx, scope, exec)
	}
}

func (spec *ReferenceActionSpec) loadAction(ctx context.Context, s *scribe.Scribe, store gitstore.Store) (ActionSpec, error) {
	span := trace.SpanFromContext(ctx)

	// 1. First, try reading "action.yml" or "action.yaml" file
	for _, f := range []string{"action.yml", "action.yaml"} {
		path := filepath.Join(spec.Repo.Path, f)
		if r, err := store.File(ctx, spec.Repo, spec.rev, path); err == nil {
			span.AddEvent("Loaded Action",
				trace.WithAttributes(xotel.ActionPath(path)),
			)
			s.Debugf("Loading %q for action", path)
			return spec.loadActionManifest(r)
		} else if !errors.Is(err, object.ErrFileNotFound) {
			return nil, err
		}
	}

	// 2. Second, try reading "Dockerfile" or "dockerfile"
	for _, f := range []string{"Dockerfile", "dockerfile"} {
		path := filepath.Join(spec.Repo.Path, f)
		if r, err := store.File(ctx, spec.Repo, spec.rev, path); err == nil {
			r.Close()
			span.AddEvent("Loaded Action",
				trace.WithAttributes(xotel.ActionPath(path)),
			)
			s.Debugf("Loading %q for action", path)
			return spec.createDockerfileAction(path)
		} else if !errors.Is(err, object.ErrFileNotFound) {
			return nil, err
		}
	}

	return nil, fmt.Errorf(`file "action.yml", "action.yaml" and "Dockerfile" not found in your given path`)
}

func (spec *ReferenceActionSpec) loadActionManifest(r io.ReadCloser) (ActionSpec, error) {
	defer r.Close()

	m := make(map[string]any)
	if err := yaml.NewDecoder(r).Decode(m); err != nil {
		return nil, err
	}
	action := new(actions.Action)
	if err := model.Decode(m, action); err != nil {
		return nil, err
	}

	return ToActionSpec(action, spec.Repo)
}

func (spec *ReferenceActionSpec) createDockerfileAction(dockerfile string) (ActionSpec, error) {
	action := &DockerActionSpec{
		Image: dockerfile,
	}
	return action, nil
}

func (spec *ReferenceActionSpec) transferAction(ctx context.Context, store gitstore.Store, sandbox sandboxer.Sandbox) error {
	location := repository.FullName(spec.Repo) + "@" + spec.Repo.Ref
	r, err := store.Read(ctx, spec.Repo, spec.rev, location)
	if err != nil {
		return err
	}
	defer r.Close()

	return sandbox.CopyIn(ctx, r, sandbox.Layout().Actions)
}

func (spec *ReferenceActionSpec) serverDomain(s string) string {
	if u, err := url.Parse(s); err == nil {
		return u.Host
	}
	return ""
}
