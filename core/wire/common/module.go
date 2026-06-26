/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_common

import (
	"fmt"
	"strconv"

	exec "drassi.run/core/pkg/executor"
	expr "drassi.run/core/pkg/expression"
	libs "drassi.run/core/pkg/expression/libraries"
	"drassi.run/core/pkg/feature"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/problem"
	xdig "drassi.run/core/util/dig"
	"drassi.run/core/wire"
	"go.uber.org/dig"
)

type Option func(o *options)
type options struct {
	defaultDossier      bool          // create default records.Dossier
	defaultFeatureFlags bool          // use feature.Empty as feature.Flags
	expressionOptions   []expr.Option // additional expression.Option
}

func ProvideDefaultDossier(b bool) Option {
	return func(o *options) {
		o.defaultDossier = b
	}
}

func UseEmptyFeatureFlags(b bool) Option {
	return func(o *options) {
		o.defaultFeatureFlags = b
	}
}

func AddExpressionOption(opts ...expr.Option) Option {
	return func(o *options) {
		o.expressionOptions = append(o.expressionOptions, opts...)
	}
}

func Module(opts ...Option) *wire.Module {
	o := &options{
		defaultDossier:      true,
		defaultFeatureFlags: true,
	}
	for _, opt := range opts {
		opt(o)
	}

	fn := func(scope *dig.Scope) error {
		if err := xdig.Supply(scope, exec.CIEnv, dig.Group(wire.EnvProvider)); err != nil {
			return fmt.Errorf("provide CIEnv: %w", err)
		}
		if err := xdig.Supply[exec.EnvProvider](scope, new(githubEnv), dig.Group(wire.EnvProvider)); err != nil {
			return fmt.Errorf("provide 'github' EnvProvider: %w", err)
		}
		if err := scope.Provide(runnerEnv, dig.Group(wire.EnvProvider)); err != nil {
			return fmt.Errorf("provide 'runner' EnvProvider: %w", err)
		}

		if err := scope.Provide(problem.NewMatchers); err != nil {
			return fmt.Errorf("provide problem.Matchers: %w", err)
		}

		if o.defaultDossier {
			if err := scope.Provide(newDossier); err != nil {
				return fmt.Errorf("provide default records.Dossier: %w", err)
			}
		}

		if err := scope.Provide(getGitHub); err != nil {
			return fmt.Errorf("provide records.GitHub: %w", err)
		}

		if err := scope.Provide(getEnv); err != nil {
			return fmt.Errorf("provide Env: %w", err)
		}

		if err := scope.Provide(o.expressionEnv); err != nil {
			return fmt.Errorf("provide default expression.Env: %w", err)
		}

		if o.defaultFeatureFlags {
			if err := xdig.Supply(scope, feature.Empty); err != nil {
				return fmt.Errorf("provide default (empty) feature.Flags: %w", err)
			}
		}

		if err := scope.Provide(collectEnvProvider); err != nil {
			return fmt.Errorf("collect EnvProviders: %w", err)
		}
		if err := scope.Provide(collectPostStartHook[exec.JobExecutor], dig.Name(wire.PostStart)); err != nil {
			return fmt.Errorf("collect 'post-start' Hooks: %w", err)
		}
		if err := scope.Provide(collectPreStopHook[exec.JobExecutor], dig.Name(wire.PreStop)); err != nil {
			return fmt.Errorf("collect 'pre-stop' Hooks: %w", err)
		}

		return nil
	}
	return wire.NewModule("core/common", fn)
}

func (o *options) expressionEnv(d *records.Dossier) (expr.Env, error) {
	opts := []expr.Option{
		expr.WithCache(true),
		expr.WithLibrary(libs.StdLib()),
		expr.WithVariable("secrets", d.Secrets),
		expr.WithVariable("vars", d.Variables),
		expr.WithVariable("needs", d.Needs),
		expr.WithVariable("strategy", d.Strategy),
		expr.WithVariable("matrix", d.Matrix),
		expr.WithVariable("inputs", d.Inputs),
	}
	opts = append(opts, o.expressionOptions...)
	return expr.NewEnv(opts...)
}

func newDossier(runner *records.Runner) *records.Dossier {
	d := new(records.Dossier)
	d.Github = new(records.Github)
	d.Env = make(map[string]string)
	d.Runner = runner
	return d
}

func getGitHub(d *records.Dossier) *records.Github {
	return d.Github
}

func getEnv(d *records.Dossier) map[string]string {
	return d.Env
}

type githubEnv struct{}

func (g *githubEnv) Env(e exec.StepExecutor) map[string]string {
	gh := e.Github()

	// set GITHUB_* env
	return map[string]string{
		"GITHUB_ACTION":              gh.Action,
		"GITHUB_ACTION_REF":          gh.ActionRef,
		"GITHUB_ACTION_REPOSITORY":   gh.ActionRepository,
		"GITHUB_ACTOR":               gh.Actor,
		"GITHUB_ACTOR_ID":            gh.ActorId,
		"GITHUB_API_URL":             gh.ApiUrl,
		"GITHUB_BASE_REF":            gh.BaseRef,
		"GITHUB_EVENT_NAME":          gh.EventName,
		"GITHUB_EVENT_PATH":          gh.EventPath,
		"GITHUB_GRAPHQL_URL":         gh.GraphqlUrl,
		"GITHUB_HEAD_REF":            gh.HeadRef,
		"GITHUB_JOB":                 gh.Job,
		"GITHUB_REF":                 gh.Ref,
		"GITHUB_REF_NAME":            gh.RefName,
		"GITHUB_REF_PROTECTED":       strconv.FormatBool(gh.RefProtected),
		"GITHUB_REF_TYPE":            string(gh.RefType),
		"GITHUB_REPOSITORY":          gh.Repository,
		"GITHUB_REPOSITORY_ID":       gh.RepositoryId,
		"GITHUB_REPOSITORY_OWNER":    gh.RepositoryOwner,
		"GITHUB_REPOSITORY_OWNER_ID": gh.RepositoryOwnerId,
		"GITHUB_RETENTION_DAYS":      gh.RetentionDays,
		"GITHUB_RUN_ATTEMPT":         gh.RunAttempt,
		"GITHUB_RUN_ID":              gh.RunId,
		"GITHUB_RUN_NUMBER":          gh.RunNumber,
		"GITHUB_SERVER_URL":          gh.ServerUrl,
		"GITHUB_SHA":                 gh.Sha,
		"GITHUB_TRIGGERING_ACTOR":    gh.TriggeringActor,
		"GITHUB_WORKFLOW":            gh.Workflow,
		"GITHUB_WORKFLOW_REF":        gh.WorkflowRef,
		"GITHUB_WORKFLOW_SHA":        gh.WorkflowSha,
		"GITHUB_WORKSPACE":           gh.Workspace,
	}
}

func runnerEnv(runner *records.Runner) exec.EnvProvider {
	m := map[string]string{
		"RUNNER_NAME":        runner.Name,
		"RUNNER_ARCH":        string(runner.Arch),
		"RUNNER_OS":          string(runner.Os),
		"RUNNER_ENVIRONMENT": runner.Environment,
		"RUNNER_TEMP":        runner.Temp,
		"RUNNER_TOOL_CACHE":  runner.ToolCache,
		"RUNNER_WORKSPACE":   runner.Workspace,
	}
	if runner.Debug == "1" {
		m["RUNNER_DEBUG"] = "1"
	}
	return exec.StaticEnv(m)
}
