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
	"net/http"
	"regexp"
	"slices"
	"strings"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/problem"
	"drassi.run/core/pkg/executor/secret"
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/libraries"
	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/pkg/stream"
	"drassi.run/core/util/dig"
	"drassi.run/core/wire/cmdhandler"
	"drassi.run/core/wire/etc"
	"drassi.run/core/wire/runtime"
	"drassi.run/core/wire/streams"
	"drassi.run/gha-runner/pkg/lease"
	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/report"
	"drassi.run/gha-runner/pkg/reporter"
	"drassi.run/gha-runner/pkg/types"
	"go.uber.org/dig"
	"golang.org/x/oauth2"
)

type Worker struct {
	msg       *messages.PipelineAgentJobRequest
	runnerSvc *lease.RunnerService

	ctx    context.Context
	cancel context.CancelCauseFunc

	lease    lease.Lease
	reporter *reporter.GhaReporter

	exec     executor.JobExecutor
	cleaners []func(ctx context.Context) error
	waiters  []types.Waiter
}

func New(msg *messages.PipelineAgentJobRequest) *Worker {
	return &Worker{msg: msg}
}

func (w *Worker) setup(scope *dig.Scope) error {
	if err := w.initService(scope); err != nil {
		return err
	}
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

func (w *Worker) initService(scope *dig.Scope) error {
	if err := xdig.Supply(scope, w.runnerSvc); err != nil {
		return err
	}

	ep := w.msg.ServiceEndpoint("SystemVssConnection")
	if ep == nil {
		return fmt.Errorf("SystemVssConnection service endpoint not available")
	}

	hc := http.DefaultClient
	if source, err := ep.TokenSource(); err != nil {
		return err
	} else if source != nil {
		hc = oauth2.NewClient(w.ctx, source)
	}

	switch typ := w.msg.MessageType; typ {
	case messages.TypeRunnerJobRequest:
		return w.initRunnerService(ep, hc, scope)
	case messages.TypePipelineAgentJobRequest:
		return w.initPipelineAgentService(ep, hc, scope)
	default:
		return fmt.Errorf("PipelineAgentJobRequest - unknown message type: %s", typ)
	}
}

// init services used to handle MessageType="RunnerJobRequest"
func (w *Worker) initRunnerService(ep *messages.ServiceEndpoint, hc *http.Client, scope *dig.Scope) error {
	if svc, err := lease.NewRunService(ep.Url, hc); err != nil {
		return err
	} else {
		if err = xdig.Supply(scope, svc); err != nil {
			return err
		}

		w.lease = svc.Lease(w.msg)
	}

	if url := ep.Data["ResultsServiceUrl"]; url != "" {
		if svc, err := report.NewResultService(url, hc, w.msg); err != nil {
			return err
		} else {
			if err = xdig.Supply(scope, svc); err != nil {
				return err
			}
			if err = provideResultServiceSubscribers(scope); err != nil {
				return err
			}

			recorder := svc.TimelineRecorder()
			w.reporter = reporter.New(recorder)
		}
	}

	return nil
}

// init services used to handle MessageType="PipelineAgentJobRequest"
func (w *Worker) initPipelineAgentService(ep *messages.ServiceEndpoint, hc *http.Client, scope *dig.Scope) error {
	w.lease = w.runnerSvc.Lease(w.msg)

	if svc, err := report.NewJobService(ep.Url, hc, w.msg); err != nil {
		return err
	} else {
		if err = xdig.Supply(scope, svc); err != nil {
			return err
		}
		if err = provideJobServiceSubscribers(scope); err != nil {
			return err
		}

		w.lease = svc.WrapLease(w.lease)

		recorder := svc.TimelineRecorder()
		w.reporter = reporter.New(recorder)
	}

	return nil
}

func (w *Worker) initScope(scope *dig.Scope) error {
	req := w.lease.GetMessage()

	// expression.Env
	dossier := new(records.Dossier)
	if err := model.Decode(req.ContextData, dossier); err != nil {
		return err
	}
	opts := []expression.Option{
		expression.WithCache(true),
		expression.WithLibrary(libraries.StdLib()),
		expression.WithVariable("secrets", dossier.Secrets),
		expression.WithVariable("vars", dossier.Variables),
		expression.WithVariable("needs", dossier.Needs),
		expression.WithVariable("strategy", dossier.Strategy),
		expression.WithVariable("matrix", dossier.Matrix),
		expression.WithVariable("inputs", dossier.Inputs),
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
	// https://github.com/actions/runner/blob/v2.324.0/src/Runner.Worker/ExecutionContext.cs#L882-L891
	github := dossier.Github
	if github.Token == "" {
		if v, ok := req.Variables["system.github.token"]; ok {
			github.Token = v.Value
		} else if v, ok = req.Variables["github_token"]; ok {
			github.Token = v.Value
		}
	}
	if github.Job == "" {
		if v, ok := req.Variables["system.github.job"]; ok {
			github.Job = v.Value
		}
	}
	if err := xdig.Supply(scope, *github); err != nil {
		return err
	}

	// secret.Masker
	// https://github.com/actions/runner/blob/v2.323.0/src/Runner.Worker/Worker.cs#L140
	sm := secret.NewMasker()
	for _, v := range req.Variables {
		if v.IsSecret {
			sm.AddSecret(secret.NewValueSecret(v.Value))
		}
	}
	for _, s := range req.MaskHints {
		switch s.Type {
		case messages.MaskTypeVariable:
			sm.AddSecret(secret.NewValueSecret(s.Value))
		case messages.MaskTypeRegex:
			if re, err := regexp.Compile(s.Value); err != nil {
				return fmt.Errorf("invalid regex %q: %w", s.Value, err)
			} else {
				sm.AddSecret(secret.NewRegexSecret(re))
			}
		default:
			return fmt.Errorf("unknown mask type %q", s.Type)
		}
	}
	if res := req.Resources; res != nil {
		for _, ep := range res.Endpoints {
			if authz := ep.Authorization; authz != nil {
				for _, v := range authz.Parameters {
					if v != "" {
						sm.AddSecret(secret.NewValueSecret(v))
					}
				}
			}
		}
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
	if err := wire_runtime.ProvideTo(scope); err != nil {
		return err
	}
	if err := etc.Wire(scope); err != nil {
		return err
	}

	// https://github.com/actions/runner/blob/v2.323.0/src/Runner.Worker/Handlers/NodeScriptActionHandler.cs#L53-L78
	// https://github.com/actions/runner/blob/v2.323.0/src/Runner.Worker/Handlers/ContainerActionHandler.cs#L218-L238
	sysCon := req.ServiceEndpoint("SystemVssConnection")
	if sysCon == nil {
		return fmt.Errorf("service endpoint 'SystemVssConnection' not found")
	}
	var accessToken string
	if authz := sysCon.Authorization; authz != nil && authz.Scheme == "OAuth" {
		accessToken = authz.Parameters["AccessToken"]
	}
	sysEnv := map[string]string{
		"GITHUB_ACTIONS":        "true",
		"ACTIONS_RUNTIME_URL":   sysCon.Url,
		"ACTIONS_RUNTIME_TOKEN": accessToken,
	}
	if url := sysCon.Data["CacheServerUrl"]; url != "" {
		sysEnv["ACTIONS_CACHE_URL"] = url
	}
	if cacheV2 := req.Variables["actions_uses_cache_service_v2"]; strings.ToLower(cacheV2.Value) == "true" {
		sysEnv["ACTIONS_CACHE_SERVICE_V2"] = "True" // bool.TrueString
	}
	if url := sysCon.Data["PipelinesServiceUrl"]; url != "" {
		sysEnv["ACTIONS_RUNTIME_URL"] = url
	}
	if url := sysCon.Data["GenerateIdTokenUrl"]; url != "" {
		sysEnv["ACTIONS_ID_TOKEN_REQUEST_URL"] = url
		sysEnv["ACTIONS_ID_TOKEN_REQUEST_TOKEN"] = accessToken
	}
	if url := sysCon.Data["ResultsServiceUrl"]; url != "" {
		sysEnv["ACTIONS_RESULTS_URL"] = url
	} else if v, ok := req.Variables["system.github.results_endpoint"]; ok {
		sysEnv["ACTIONS_RESULTS_URL"] = v.Value
	}

	//hc := http.DefaultClient
	//if source, err := sysCon.TokenSource(); err == nil && source != nil {
	//	hc = oauth2.NewClient(w.ctx, source)
	//}
	//
	//cp := xcontext.NewStaticProvider(w.ctx)
	//if url, ok := sysCon.Data["FeedStreamUrl"]; ok && url != "" {
	//	if log, err := service.NewConsoleLiveFeeder(cp, url, hc); err != nil {
	//		return err
	//	} else if err = xdig.Supply[stream.Handler](scope, log); err != nil {
	//		return err
	//	} else if err = log.Start(); err != nil {
	//		return err
	//	} else {
	//		w.addCleaner(log.Close)
	//	}
	//}
	//
	//rep := service.New(w.task.Id, client, cp, log, w.Cancel)
	//if err := xdig.Supply[reporter.Reporter](scope, rep); err != nil {
	//	return err
	//}
	//if err := rep.Start(); err != nil {
	//	return err
	//}
	//w.addCleaner(rep.Close)

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

func (w *Worker) initExecutor(scope *dig.Scope) error {
	req := w.lease.GetMessage()

	jr, err := messages.ToJobRun(req)
	if err != nil {
		return err
	}

	w.exec = executor.NewJobExecutor(jr)
	w.addCleanerContext(w.exec.Finalize)
	scope = scope.Scope(fmt.Sprintf("job(%s)", executor.JobId(w.exec)))

	return w.exec.Initialize(w.ctx, scope)
}

func (w *Worker) Run(ctx context.Context, scope *dig.Scope) (err error) {
	defer w.wait()

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
	go w.lease.Renew(ctx)
	defer func() {
		record := w.reporter.JobRecord()
		ex := w.lease.Complete(w.ctx, record)
		err = errors.Join(err, ex)
	}()

	r := w.exec.RunJob(w.ctx)
	if r.Result != records.ResultSuccess {
		return fmt.Errorf("job failed")
	}
	return nil
}

func (w *Worker) wait() {
	for _, waiter := range w.waiters {
		waiter.Wait()
	}
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
