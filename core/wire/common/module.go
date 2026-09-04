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
	expressionAstCache  bool          // enable caching expression AST
	forgeEnvPrefix      string
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

func EnableExpressionAstCache() Option {
	return func(o *options) {
		o.expressionAstCache = true
	}
}

func ForgeEnvPrefix(s string) Option {
	return func(o *options) {
		o.forgeEnvPrefix = s
	}
}

func Module(opts ...Option) *wire.Module {
	o := &options{
		defaultDossier:      true,
		defaultFeatureFlags: true,
		forgeEnvPrefix:      "GITHUB_",
	}
	for _, opt := range opts {
		opt(o)
	}

	fn := func(scope *dig.Scope) error {
		if err := xdig.Supply(scope, exec.CIEnv, dig.Group(wire.EnvProvider)); err != nil {
			return fmt.Errorf("provide CIEnv: %w", err)
		}
		if err := scope.Provide(StateEnv, dig.Group(wire.EnvProvider)); err != nil {
			return fmt.Errorf("provide 'STATE_' EnvProvider: %w", err)
		}
		if err := scope.Provide(o.ForgeEnv, dig.Group(wire.EnvProvider)); err != nil {
			return fmt.Errorf("provide 'GITHUB_' EnvProvider: %w", err)
		}
		if err := scope.Provide(RunnerEnv, dig.Group(wire.EnvProvider)); err != nil {
			return fmt.Errorf("provide 'RUNNER_' EnvProvider: %w", err)
		}

		if err := scope.Provide(problem.NewDetectorFactory); err != nil {
			return fmt.Errorf("provide problem.DetectorFactory: %w", err)
		}

		if o.defaultDossier {
			if err := scope.Provide(newDossier); err != nil {
				return fmt.Errorf("provide default records.Dossier: %w", err)
			}
		}

		if err := scope.Provide(getForge); err != nil {
			return fmt.Errorf("provide records.Forge: %w", err)
		}

		if err := scope.Provide(getEnv); err != nil {
			return fmt.Errorf("provide Env: %w", err)
		}

		if err := scope.Provide(o.expressionEnv); err != nil {
			return fmt.Errorf("provide default expression.Env: %w", err)
		}

		if o.expressionAstCache {
			if err := scope.Decorate(expr.EnableCache); err != nil {
				return fmt.Errorf("enable expression AST cache: %w", err)
			}
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

func newDossier(runner *records.RunnerInfo) *records.Dossier {
	d := new(records.Dossier)
	d.Forge = new(records.Forge)
	d.Env = make(map[string]string)
	d.Runner = runner
	return d
}

func getForge(d *records.Dossier) *records.Forge {
	return d.Forge
}

func getEnv(d *records.Dossier) map[string]string {
	return d.Env
}

func StateEnv() exec.EnvProvider {
	fn := func(e exec.StepExecutor) map[string]string {
		m := make(map[string]string)

		// set STATE_* env
		// https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#sending-values-to-the-pre-and-post-actions
		for k, v := range e.State() {
			k = "STATE_" + k
			m[k] = v
		}
		return m
	}
	return exec.EnvProviderFunc(fn)
}

func (o *options) ForgeEnv() exec.EnvProvider {
	fn := func(e exec.StepExecutor) map[string]string {
		forge := e.Forge()

		// set GITHUB_* env
		return map[string]string{
			o.forgeEnvPrefix + "ACTION":              forge.Action,
			o.forgeEnvPrefix + "ACTION_REF":          forge.ActionRef,
			o.forgeEnvPrefix + "ACTION_REPOSITORY":   forge.ActionRepository,
			o.forgeEnvPrefix + "ACTOR":               forge.Actor,
			o.forgeEnvPrefix + "ACTOR_ID":            forge.ActorId,
			o.forgeEnvPrefix + "API_URL":             forge.ApiUrl,
			o.forgeEnvPrefix + "BASE_REF":            forge.BaseRef,
			o.forgeEnvPrefix + "EVENT_NAME":          forge.EventName,
			o.forgeEnvPrefix + "EVENT_PATH":          forge.EventPath,
			o.forgeEnvPrefix + "GRAPHQL_URL":         forge.GraphqlUrl,
			o.forgeEnvPrefix + "HEAD_REF":            forge.HeadRef,
			o.forgeEnvPrefix + "JOB":                 forge.Job,
			o.forgeEnvPrefix + "REF":                 forge.Ref,
			o.forgeEnvPrefix + "REF_NAME":            forge.RefName,
			o.forgeEnvPrefix + "REF_PROTECTED":       strconv.FormatBool(forge.RefProtected),
			o.forgeEnvPrefix + "REF_TYPE":            string(forge.RefType),
			o.forgeEnvPrefix + "REPOSITORY":          forge.Repository,
			o.forgeEnvPrefix + "REPOSITORY_ID":       forge.RepositoryId,
			o.forgeEnvPrefix + "REPOSITORY_OWNER":    forge.RepositoryOwner,
			o.forgeEnvPrefix + "REPOSITORY_OWNER_ID": forge.RepositoryOwnerId,
			o.forgeEnvPrefix + "RETENTION_DAYS":      forge.RetentionDays,
			o.forgeEnvPrefix + "RUN_ATTEMPT":         forge.RunAttempt,
			o.forgeEnvPrefix + "RUN_ID":              forge.RunId,
			o.forgeEnvPrefix + "RUN_NUMBER":          forge.RunNumber,
			o.forgeEnvPrefix + "SERVER_URL":          forge.ServerUrl,
			o.forgeEnvPrefix + "SHA":                 forge.Sha,
			o.forgeEnvPrefix + "TRIGGERING_ACTOR":    forge.TriggeringActor,
			o.forgeEnvPrefix + "WORKFLOW":            forge.Workflow,
			o.forgeEnvPrefix + "WORKFLOW_REF":        forge.WorkflowRef,
			o.forgeEnvPrefix + "WORKFLOW_SHA":        forge.WorkflowSha,
			o.forgeEnvPrefix + "WORKSPACE":           forge.Workspace,
		}
	}
	return exec.EnvProviderFunc(fn)
}

func RunnerEnv(runner *records.RunnerInfo) exec.EnvProvider {
	fn := func(e exec.StepExecutor) map[string]string {
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
		return m
	}
	return exec.EnvProviderFunc(fn)
}
