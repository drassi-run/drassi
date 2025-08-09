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
	"io"
	"slices"
	"strings"

	runnerv1 "code.gitea.io/actions-proto-go/runner/v1"
	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/problem"
	"drassi.run/core/pkg/executor/runtime"
	"drassi.run/core/pkg/executor/secret"
	"drassi.run/core/pkg/executor/support"
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/libraries"
	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/pkg/stream"
	"drassi.run/core/pkg/wire/cmdhandler"
	"drassi.run/core/pkg/wire/etc"
	"drassi.run/core/pkg/wire/runtime"
	"drassi.run/core/pkg/wire/streams"
	"drassi.run/core/util/context"
	"drassi.run/core/util/dig"
	"drassi.run/gitea-runner/pkg/reporter"
	"drassi.run/gitea-runner/pkg/service"
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
	if err := w.initContext(scope); err != nil {
		return err
	}
	if err := w.initExecutor(scope); err != nil {
		return err
	}

	return nil
}

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
	pm := make(map[string]problem.Matcher)
	if err := xdig.Supply(scope, pm); err != nil {
		return err
	}

	// Wire scope
	if err := scope.Provide(command.NewFileManager); err != nil {
		return err
	}
	if err := scope.Provide(command.NewConsoleManager); err != nil {
		return err
	}
	if err := wire_cmdhandler.ProvideTo(scope); err != nil {
		return err
	}
	if err := wire_streams.ProvideTo(scope.Scope("internal(streams)")); err != nil {
		return err
	}
	if err := scope.Provide(newContainerRuntime(w.ctx, &github)); err != nil {
		return err
	}

	if err := etc.Wire(scope); err != nil {
		return err
	}

	var client service.GiteaClient
	if err := xdig.Populate(scope, &client); err != nil {
		return err
	}

	cp := xcontext.NewStaticProvider(w.ctx)

	log := reporter.NewLogStreamer(w.task.Id, cp, client)
	if err := xdig.Supply[stream.Handler](scope, log); err != nil {
		return err
	}
	if err := log.Start(); err != nil {
		return err
	}
	w.addCleaner(log.Close)

	rep := reporter.New(w.task.Id, client, cp, log, w.Cancel)
	if err := xdig.Supply(scope, rep); err != nil {
		return err
	}
	if err := rep.Start(); err != nil {
		return err
	}
	w.addCleaner(rep.Close)

	if err := scope.Provide(reporter.NewListener,
		dig.As(new(executor.JobListener), new(executor.StepListener)),
		dig.Name("reporter")); err != nil {
		return err
	}

	if err := scope.Provide(composeJobListener); err != nil {
		return err
	}
	if err := scope.Provide(composeStepListener); err != nil {
		return err
	}

	if err := scope.Invoke(w.provideEnv); err != nil {
		return err
	}

	return scope.Invoke(func(streams stream.Streams) {
		if closer, ok := streams.Out().(io.Closer); ok {
			w.addCleaner(closer.Close)
		}
		if closer, ok := streams.Err().(io.Closer); ok {
			w.addCleaner(closer.Close)
		}
	})
}

func (w *Worker) initContext(scope *dig.Scope) error {
	var diary scribe.Diary
	if err := xdig.Populate(scope, &diary); err != nil {
		return err
	}

	w.ctx = scribe.ContextWithScribe(w.ctx, diary)
	return nil
}

func (w *Worker) provideEnv(client service.GiteaClient, github records.Github, envProv support.EnvProvider) error {
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

	envProv.ProvideEnv(m)
	return nil
}

func (w *Worker) initExecutor(scope *dig.Scope) error {
	workflow := new(workflows.Workflow)
	if err := decodeWorkflow(w.task.WorkflowPayload, workflow); err != nil {
		return err
	}
	jr, err := convertJobRun(workflow)
	if err != nil {
		return err
	}

	w.exec = executor.NewJobExecutor(jr)
	w.addCleanerContext(w.exec.Finalize)
	scope = scope.Scope(fmt.Sprintf("job(%s)", executor.JobId(w.exec)))

	return w.exec.Initialize(w.ctx, scope)
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

	// run main execution
	r := w.exec.RunJob(w.ctx)
	if r.Result != records.ResultSuccess {
		return fmt.Errorf("job failed")
	}
	return nil
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
	container.Engine, stream.Streams, sandboxer.Sandbox, *records.JobInfo,
) (runtime.Container, error) {
	return func(
		engine container.Engine,
		streams stream.Streams,
		sandbox sandboxer.Sandbox,
		info *records.JobInfo,
	) (runtime.Container, error) {
		return wire_runtime.NewContainerRuntime(ctx, engine, streams, sandbox, info, gh)
	}
}

type listenerParams[L any] struct {
	dig.In
	Stack    L `name:"stack"`
	Reporter L `name:"reporter"`
	Command  L `name:"command"`
}

func composeJobListener(p listenerParams[executor.JobListener]) executor.JobListener {
	return executor.NewCompositeJobListener(p.Stack, p.Reporter, p.Command)
}

func composeStepListener(p listenerParams[executor.StepListener]) executor.StepListener {
	return executor.NewCompositeStepListener(p.Stack, p.Reporter, p.Command)
}
