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

// Example:
// + Using a public action: `uses: actions/aws@v2.0.1`
// + Using a public action in a subdirectory: `uses: actions/aws/ec2@main`
// + Using a local action: `uses: ./.github/actions/hello-world-action`
type ActionStepRun struct {
	BaseStepRun
	Repo *repository.Repository

	rev       string
	actionRun StepRun
}

func (sr *ActionStepRun) Repository() *repository.Repository {
	return sr.Repo
}

func (sr *ActionStepRun) Unwrap() StepRun {
	return sr.actionRun
}

func (sr *ActionStepRun) Initialize(ctx context.Context, scope *dig.Scope) error {
	s := scribe.FromContext(ctx)

	var (
		github  records.Github
		store   gitstore.Store
		sandbox sandboxer.Sandbox
		exprEnv expression.Env
	)

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(xotel.ActionRepo(repository.Location(sr.Repo)))

	if err := xdig.Populate(scope, &github); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &store); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &sandbox); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &exprEnv); err != nil {
		return err
	}
	if err := xdig.Supply(scope, sr.Repo); err != nil {
		return err
	}

	token := github.Token
	// If the action is located in different server than the job repo,
	// unset the token to prevent an unauthenticated error.
	if repository.Endpoint(sr.Repo) != sr.serverDomain(github.ServerUrl) {
		token = ""
	}

	// evaluating before loadAction so it DisplayName can be push-down into child StepRun
	if err := sr.evaluateDisplayName(ctx, exprEnv, repository.Location(sr.Repo)); err != nil {
		return err
	}

	if rev, err := store.Fetch(ctx, sr.Repo, token); err != nil {
		return err
	} else {
		s.Writef("Download action repository %q (SHA:%s)", repository.Location(sr.Repo), rev)
		sr.rev = rev
	}

	if err := sr.loadAction(ctx, s, store); err != nil {
		return err
	}
	if err := sr.transferAction(ctx, store, sandbox); err != nil {
		return err
	}

	return sr.actionRun.Initialize(ctx, scope)
}

func (sr *ActionStepRun) PreTask() *Task {
	return sr.actionRun.PreTask()
}

func (sr *ActionStepRun) MainTask() *Task {
	return sr.actionRun.MainTask()
}

func (sr *ActionStepRun) PostTask() *Task {
	return sr.actionRun.PostTask()
}

func (sr *ActionStepRun) loadAction(ctx context.Context, s *scribe.Scribe, store gitstore.Store) error {
	span := trace.SpanFromContext(ctx)

	// 1. First, try reading "action.yml" or "action.yaml" file
	for _, f := range []string{"action.yml", "action.yaml"} {
		path := filepath.Join(sr.Repo.Path, f)
		if r, err := store.File(ctx, sr.Repo, sr.rev, path); err == nil {
			span.AddEvent("Loaded Action",
				trace.WithAttributes(xotel.ActionPath(path)),
			)
			s.Debugf("Loading %q for action %q", path, sr.Id)
			return sr.loadActionManifest(r)
		} else if !errors.Is(err, object.ErrFileNotFound) {
			return err
		}
	}

	// 2. Second, try reading "Dockerfile" or "dockerfile"
	for _, f := range []string{"Dockerfile", "dockerfile"} {
		path := filepath.Join(sr.Repo.Path, f)
		if r, err := store.File(ctx, sr.Repo, sr.rev, path); err == nil {
			r.Close()
			span.AddEvent("Loaded Action",
				trace.WithAttributes(xotel.ActionPath(path)),
			)
			s.Debugf("Loading %q for action %q", path, sr.Id)
			return sr.createDockerfileAction(path)
		} else if !errors.Is(err, object.ErrFileNotFound) {
			return err
		}
	}

	return fmt.Errorf(`file "action.yml", "action.yaml" and "Dockerfile" not found in your given path`)
}

func (sr *ActionStepRun) loadActionManifest(r io.ReadCloser) error {
	defer r.Close()

	m := make(map[string]any)
	if err := yaml.NewDecoder(r).Decode(m); err != nil {
		return err
	}
	action := new(actions.Action)
	if err := model.Decode(m, action); err != nil {
		return err
	}

	if actionRun, err := FromAction(action, sr.BaseStepRun); err != nil {
		return err
	} else {
		sr.actionRun = WithTelemetryStepRun(actionRun)
	}

	return nil
}

func (sr *ActionStepRun) createDockerfileAction(dockerfile string) error {
	actionRun := &DockerStepRun{
		BaseStepRun: sr.BaseStepRun,
		Image:       dockerfile,
	}
	sr.actionRun = WithTelemetryStepRun(actionRun)
	return nil
}

func (sr *ActionStepRun) transferAction(ctx context.Context, store gitstore.Store, sandbox sandboxer.Sandbox) error {
	location := repository.FullName(sr.Repo) + "@" + sr.Repo.Ref
	r, err := store.Read(ctx, sr.Repo, sr.rev, location)
	if err != nil {
		return err
	}
	defer r.Close()

	return sandbox.CopyIn(ctx, r, sandbox.Layout().Actions)
}

func (sr *ActionStepRun) serverDomain(s string) string {
	if u, err := url.Parse(s); err == nil {
		return u.Host
	}
	return ""
}
