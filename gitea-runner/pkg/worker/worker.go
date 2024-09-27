package worker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"

	runnerv1 "code.gitea.io/actions-proto-go/runner/v1"
	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/problem"
	"drassi.run/core/pkg/executor/reporter"
	"drassi.run/core/pkg/executor/secret"
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/libraries"
	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/util/dig"
	"drassi.run/core/pkg/wire/cmdhandler"
	"drassi.run/core/pkg/wire/etc"
	"drassi.run/core/pkg/wire/reporter"
	"drassi.run/core/pkg/wire/streams"
	"drassi.run/gitea-runner/pkg/service"
	"go.uber.org/dig"
)

type Worker struct {
	ctx  context.Context
	task *runnerv1.Task

	exec     executor.JobExecutor
	cleaners []func() error
}

func New(ctx context.Context, task *runnerv1.Task) *Worker {
	return &Worker{
		ctx:  ctx,
		task: task,
	}
}

func (w *Worker) Setup(scope *dig.Scope) error {
	if err := w.initScope(scope); err != nil {
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
	if err := scope.Provide(newCommandConsoleManager); err != nil {
		return err
	}
	if err := scope.Provide(executor.NewSupervisor); err != nil {
		return err
	}
	if err := wire_cmdhandler.ProvideTo(scope); err != nil {
		return err
	}
	if err := wire_streams.ProvideTo(scope.Scope("internal(streams)")); err != nil {
		return err
	}

	if err := etc.Wire(scope); err != nil {
		return err
	}

	var client service.GiteaClient
	if err := xdig.Populate(scope, &client); err != nil {
		return err
	}
	rep := service.NewReporter(w.ctx, w.task.Id, client)
	w.addCleaner(rep.Close)

	if err := xdig.Supply[reporter.Reporter](scope, rep); err != nil {
		return err
	}
	if err := wire_reporter.ProvideTo(scope); err != nil {
		return err
	}
	if err := wire_reporter.Wire(scope); err != nil {
		return err
	}

	return scope.Invoke(func(streams *sandboxer.Streams) {
		if closer, ok := streams.Out.(io.Closer); ok {
			w.addCleaner(closer.Close)
		}
		if closer, ok := streams.Err.(io.Closer); ok {
			w.addCleaner(closer.Close)
		}
	})
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
	w.addCleaner(w.exec.Finalize)

	scope = scope.Scope(fmt.Sprintf("job(%s)", executor.JobId(w.exec)))
	w.exec.SetContext(w.ctx)

	return w.exec.Initialize(scope)
}

func (w *Worker) Run() error {
	r := w.exec.RunJob()
	if r.Result != records.ResultSuccess {
		return fmt.Errorf("job failed")
	}
	return nil
}

func (w *Worker) Teardown() error {
	errs := make([]error, 0)
	for _, cleaner := range slices.Backward(w.cleaners) {
		errs = append(errs, cleaner())
	}
	return errors.Join(errs...)
}

func (w *Worker) addCleaner(c func() error) {
	w.cleaners = append(w.cleaners, c)
}

type cmParams struct {
	dig.In
	StdOut io.Writer `name:"stdout"`
}

func newCommandConsoleManager(p cmParams) command.ConsoleManager {
	return command.NewConsoleManager(p.StdOut)
}
