/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package worker

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	runnerv1 "code.gitea.io/actions-proto-go/runner/v1"
	"drassi.run/core/pkg/command/cmdtypes"
	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/libraries"
	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/problem"
	"drassi.run/core/pkg/runtime"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/pkg/secret"
	"drassi.run/core/pkg/stream"
	"drassi.run/core/util/context"
	"drassi.run/core/util/dig"
	"drassi.run/core/wire"
	"drassi.run/core/wire/command"
	"drassi.run/core/wire/runtime"
	"drassi.run/core/wire/streams"
	"drassi.run/core/wire/support"
	"drassi.run/gitea-runner/pkg/gitea"
	"drassi.run/gitea-runner/pkg/reporter"
	"go.uber.org/dig"
)

type Worker struct {
	task *runnerv1.Task

	ctx      context.Context
	cancel   context.CancelCauseFunc
	exec     executor.JobExecutor
	cleaners []func(ctx context.Context) error
}

func New(task *runnerv1.Task) *Worker {
	return &Worker{task: task}
}

func (w *Worker) setup(scope *dig.Scope) error {
	if err := w.initScope(scope); err != nil {
		return err
	}
	return w.initContext(scope)
}

func (w *Worker) Context() context.Context {
	return w.ctx
}

//goland:noinspection GoResourceLeak
func (w *Worker) initScope(scope *dig.Scope) error {
	// expression.Env
	needs := convertJobNeeds(w.task.Needs)
	opts := []expression.Option{
		expression.WithCache(true),
		expression.WithLibrary(libraries.StdLib()),
		expression.WithVariable("secrets", w.task.Secrets),
		expression.WithVariable("vars", w.task.Vars),
		expression.WithVariable("needs", needs),
		expression.WithAlias("gitea", "github"), // make `gitea` variable alias to `github`
		expression.WithVariable("strategy", new(records.Strategy)),
		expression.WithVariable("matrix", make(map[string]string)),
		expression.WithVariable("inputs", make(map[string]any)),
	}
	if exprEnv, err := expression.NewEnv(opts...); err != nil {
		return err
	} else if err = xdig.Supply(scope, exprEnv); err != nil {
		return err
	}

	// Env context
	env := make(map[string]string)
	if err := xdig.Supply(scope, env); err != nil {
		return err
	}

	// GitHub context
	var github records.Github
	if err := model.Decode(w.task.Context.AsMap(), &github); err != nil {
		return err
	} else if github.Token == "" {
		if t := w.task.Secrets["GITEA_TOKEN"]; t != "" {
			github.Token = t
		} else if t = w.task.Secrets["GITHUB_TOKEN"]; t != "" {
			github.Token = t
		}
	} else if err = xdig.Supply(scope, github); err != nil {
		return err
	}

	// secret.Masker
	sm := secret.NewMasker()
	for _, v := range w.task.Secrets {
		sm.AddSecret(secret.NewValueSecret(v))
	}
	if err := xdig.Supply(scope, sm); err != nil {
		return err
	}

	// problem.Matcher
	pm := make(problem.Matchers)
	if err := xdig.Supply(scope, pm); err != nil {
		return err
	}

	// Wire scope
	if err := xdig.Supply[xcontext.Provider](scope, w); err != nil {
		return err
	}
	if err := wire_command.ProvideTo(scope); err != nil {
		return err
	}
	if err := wire_streams.ProvideTo(scope); err != nil {
		return err
	}
	if err := wire_support.Wire(scope); err != nil {
		return err
	}
	if err := scope.Provide(newContainerRuntime(w.ctx, &github)); err != nil {
		return err
	}

	var client gitea.Client
	if err := xdig.Populate(scope, &client); err != nil {
		return err
	}

	log := reporter.NewLogStreamer(w.task.Id, w, client)
	if err := xdig.Supply(scope, log, dig.As(new(stream.Handler))); err != nil {
		return err
	}
	if err := xdig.Supply[scribe.Handler](scope, log.ContextHandle); err != nil {
		return err
	}
	if err := scope.Provide(stream.NewDetachResourceHandler[executor.Milieu]); err != nil {
		return err
	}
	if err := scope.Provide(cmdtypes.Discard[executor.Milieu]); err != nil {
		return err
	}
	if err := log.Start(); err != nil {
		return err
	}
	w.addCleaner(log.Close)

	rep := reporter.New(w.task.Id, client, w, log, w.Cancel)
	if err := xdig.Supply(scope, rep,
		dig.As(new(executor.JobRunDecorator), new(executor.StepRunDecorator)),
		dig.Name("reporter"),
	); err != nil {
		return err
	}
	if err := rep.Start(); err != nil {
		return err
	}
	w.addCleaner(rep.Close)

	// decorators
	if err := scope.Provide(newJobRunDecorator); err != nil {
		return err
	}
	if err := scope.Provide(newStepRunDecorator); err != nil {
		return err
	}
	if err := scope.Provide(newActionRunDecorator); err != nil {
		return err
	}

	return scope.Provide(w.endpointEnv, dig.Group(wire.EnvProvider))
}

func (w *Worker) initContext(scope *dig.Scope) error {
	var diary scribe.Diary
	if err := xdig.Populate(scope, &diary); err != nil {
		return err
	}

	w.ctx = scribe.ContextWithScribe(w.ctx, diary)
	return nil
}

func (w *Worker) endpointEnv(client gitea.Client, github records.Github) executor.EnvProvider {
	endpoint := client.Address()
	endpoint = strings.TrimSuffix(endpoint, "/")

	taskContext := w.task.Context.Fields
	giteaRuntimeToken := taskContext["gitea_runtime_token"].GetStringValue()
	if giteaRuntimeToken == "" {
		// use task token to action api token for previous Gitea Server Versions
		giteaRuntimeToken = github.Token
	}

	m := map[string]string{
		"GITEA_ACTIONS":         "true",
		"ACTIONS_RUNTIME_URL":   endpoint + "/api/actions_pipeline/",
		"ACTIONS_RUNTIME_TOKEN": giteaRuntimeToken,
		"ACTIONS_RESULTS_URL":   endpoint,
		//"ACTIONS_CACHE_URL":     "", // TODO
	}
	return executor.StaticEnv(m)
}

func (w *Worker) Run(ctx context.Context, scope *dig.Scope) (err error) {
	w.ctx, w.cancel = context.WithCancelCause(ctx)
	defer w.cancel(nil)

	// setup & teardown worker
	defer func() {
		ex := w.teardown()
		err = errors.Join(err, ex)
	}()
	if err = w.setup(scope); err != nil {
		return err
	}

	return w.run(w.ctx, scope)
}

func (w *Worker) run(ctx context.Context, scope *dig.Scope) error {
	workflow := new(workflows.Workflow)
	if err := decodeWorkflow(w.task.WorkflowPayload, workflow); err != nil {
		return err
	}
	spec, err := convertJobSpec(workflow)
	if err != nil {
		return err
	}

	scope = scope.Scope(fmt.Sprintf("job(%s)", spec.Id))
	if job, err := executor.Run(ctx, spec, scope); err != nil {
		return err
	} else if job.Result != records.ResultSuccess {
		return fmt.Errorf("job.Result failed")
	}
	return err
}

func (w *Worker) teardown() error {
	errs := make([]error, 0)
	for _, cleaner := range slices.Backward(w.cleaners) {
		errs = append(errs, cleaner(w.ctx))
	}
	return errors.Join(errs...)
}

func (w *Worker) addCleaner(c func() error) {
	w.cleaners = append(w.cleaners, func(context.Context) error {
		return c()
	})
}

func (w *Worker) addCleanerContext(c func(ctx context.Context) error) {
	w.cleaners = append(w.cleaners, c)
}

func (w *Worker) Cancel(cause error) {
	if w.cancel != nil {
		w.cancel(cause)
	}
}

func newContainerRuntime(ctx context.Context, gh *records.Github) func(
	container.Engine, sandboxer.Sandbox, *records.JobInfo,
) (runtime.Container, error) {
	return func(
		engine container.Engine,
		sandbox sandboxer.Sandbox,
		info *records.JobInfo,
	) (runtime.Container, error) {
		return wire_runtime.NewContainerRuntime(ctx, engine, sandbox, info, gh)
	}
}

type jobRunDecoratorParam struct {
	dig.In

	Reporter executor.JobRunDecorator `name:"reporter"`
}

func newJobRunDecorator(p jobRunDecoratorParam) executor.JobRunDecorator {
	return p.Reporter
}

type stepRunDecoratorParam struct {
	dig.In

	Reporter executor.StepRunDecorator `name:"reporter"`
}

func newStepRunDecorator(p stepRunDecoratorParam) executor.StepRunDecorator {
	return p.Reporter
}

type actionRunDecoratorParam struct {
	dig.In

	ConsoleCommand executor.ActionRunDecorator `name:"command"`
}

func newActionRunDecorator(p actionRunDecoratorParam) executor.ActionRunDecorator {
	return p.ConsoleCommand
}
