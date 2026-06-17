/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package worker

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"drassi.run/core/pkg/command/cmdtypes"
	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/libraries"
	"drassi.run/core/pkg/feature"
	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/problem"
	"drassi.run/core/pkg/runtime"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/pkg/secret"
	"drassi.run/core/pkg/stream"
	xcontext "drassi.run/core/util/context"
	xdig "drassi.run/core/util/dig"
	xotel "drassi.run/core/util/otel"
	"drassi.run/core/wire"
	wire_command "drassi.run/core/wire/command"
	wire_runtime "drassi.run/core/wire/runtime"
	wire_streams "drassi.run/core/wire/streams"
	wire_support "drassi.run/core/wire/support"
	"drassi.run/gha-runner/pkg/lease"
	"drassi.run/gha-runner/pkg/log"
	"drassi.run/gha-runner/pkg/log/logsubscriber"
	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/service"
	"drassi.run/gha-runner/pkg/timeline"
	"go.uber.org/dig"
	"golang.org/x/oauth2"
)

const LogSubscribers = "log-subscribers"

func (w *Worker) setup(scope *dig.Scope) error {
	if err := w.initService(scope); err != nil {
		return err
	}
	if err := w.initManager(scope); err != nil {
		return err
	}
	if err := w.initScope(scope); err != nil {
		return err
	}
	return w.initDiary(scope)
}

func (w *Worker) initService(scope *dig.Scope) error {
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
		return w.initRunnerJobRequest(ep, hc, scope)
	case messages.TypePipelineAgentJobRequest:
		return w.initPipelineAgentJobRequest(ep, hc, scope)
	default:
		return fmt.Errorf("PipelineAgentJobRequest - unknown message type: %s", typ)
	}
}

// init services used to handle MessageType="RunnerJobRequest"
func (w *Worker) initRunnerJobRequest(ep *messages.ServiceEndpoint, hc *http.Client, scope *dig.Scope) error {
	if runSvc, err := lease.NewRunService(ep.Url, hc); err != nil {
		return err
	} else {
		w.lease = runSvc.Lease(w.msg)
		if err = xdig.Supply(scope, runSvc); err != nil {
			return err
		}
	}

	url := ep.Data["ResultsServiceUrl"]
	if url == "" {
		return nil
	}

	// Init service.ResultService
	if resultSvc, err := service.NewResultService(url, hc, w.msg); err != nil {
		return err
	} else if err = xdig.Supply(scope, resultSvc); err != nil {
		return err
	}
	// TODO: init logtypes.Appender
	//if err := scope.Provide(logsubscriber.NewLiveFeedSubscriber, dig.Group(LogSubscribers)); err != nil {
	//	return err
	//}
	if err := scope.Provide(logsubscriber.NewResultServiceStepLogsSubscriber, dig.Group(LogSubscribers)); err != nil {
		return err
	}
	return scope.Provide(service.ResultService.TimelineRecorder)
}

// init services used to handle MessageType="PipelineAgentJobRequest"
func (w *Worker) initPipelineAgentJobRequest(ep *messages.ServiceEndpoint, hc *http.Client, scope *dig.Scope) error {
	var runnerSvc *lease.RunnerService
	if err := xdig.Supply(scope, &runnerSvc); err != nil {
		return err
	}
	w.lease = runnerSvc.Lease(w.msg)

	// Init service.JobService
	if jobSvc, err := service.NewJobService(ep.Url, hc, w.msg); err != nil {
		return err
	} else {
		w.lease = jobSvc.WrapLease(w.lease)
		if err = xdig.Supply(scope, jobSvc); err != nil {
			return err
		}
	}
	if err := scope.Provide(service.JobService.LiveFeedAppender); err != nil {
		return err
	}
	if err := scope.Provide(logsubscriber.NewLiveFeedSubscriber, dig.Group(LogSubscribers)); err != nil {
		return err
	}
	if err := scope.Provide(logsubscriber.NewJobServiceLogsSubscriber, dig.Group(LogSubscribers)); err != nil {
		return err
	}
	return scope.Provide(service.JobService.TimelineRecorder)
}

func (w *Worker) initManager(scope *dig.Scope) error {
	maxLogSize := int64(100) * 1024 * 1024 // 100MiB
	//goland:noinspection GoResourceLeak
	if logMgr, err := log.NewManager("/tmp/gha-runner/", maxLogSize); err != nil {
		return fmt.Errorf("create log.Manager: %w", err)
	} else {
		w.logMgr = logMgr
		if err = xdig.Supply(scope, logMgr); err != nil {
			return fmt.Errorf("supply log.Manager: %w", err)
		}
		if err = xdig.Supply(scope, logMgr, dig.As(new(stream.Handler))); err != nil {
			return fmt.Errorf("supply log.Manager as stream.Handler: %w", err)
		}
		if err = xdig.Supply[scribe.Handler](scope, logMgr.ContextHandle); err != nil {
			return fmt.Errorf("supply log.Manager.ContextHandle() as scribe.Handler: %w", err)
		}
	}
	if err := scope.Provide(log.NewDecorator,
		dig.As(new(executor.JobRunDecorator), new(executor.StepRunDecorator)),
		dig.Name("log"),
	); err != nil {
		return fmt.Errorf("provide log.Decorator as %q JobRunDecorator & StepRunDecorator: %w", "log", err)
	}

	if err := scope.Provide(timeline.NewManager); err != nil {
		return fmt.Errorf("provide timeline.Manager: %w", err)
	}
	if err := xdig.Populate(scope, &w.timelineMgr); err != nil {
		return fmt.Errorf("populate 'timelineMgr': %w", err)
	}
	if err := xdig.Supply(scope, w.timelineMgr, dig.As(new(log.RecordStore))); err != nil {
		return fmt.Errorf("provide timeline.Manager as log.RecordStore: %w", err)
	}
	if err := xdig.Supply(scope, w.timelineMgr,
		dig.As(new(executor.JobRunDecorator), new(executor.StepRunDecorator)),
		dig.Name("timeline"),
	); err != nil {
		return fmt.Errorf("provide timeline.Manager as %q JobRunDecorator & StepRunDecorator: %w", "timeline", err)
	}

	return nil
}

func (w *Worker) initScope(scope *dig.Scope) error {
	if err := xdig.Supply(scope, w.msg); err != nil {
		return fmt.Errorf("supply PipelineAgentJobRequest: %w", err)
	}
	if err := xdig.Supply(scope, w.dossier); err != nil {
		return fmt.Errorf("supply Dossier: %w", err)
	}
	if err := xdig.Supply(scope, w.dossier.Github); err != nil {
		return fmt.Errorf("supply GitHub: %w", err)
	}

	// Env context
	env := make(map[string]string)
	if err := xdig.Supply(scope, env); err != nil {
		return err
	}

	// Wire scope
	if err := scope.Provide(printSecretSource[executor.JobExecutor], dig.Group(wire.PostStart)); err != nil {
		return err
	}
	if err := scope.Provide(printTokenPermissions[executor.JobExecutor], dig.Group(wire.PostStart)); err != nil {
		return err
	}
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
	if err := scope.Provide(w.newContainerRuntime); err != nil {
		return err
	}
	if err := scope.Provide(timeline.NewIssueReporter, dig.As(new(cmdtypes.Reporter[executor.Milieu]))); err != nil {
		return err
	}
	if err := scope.Provide(stream.NewDetachResourceHandler[executor.Milieu]); err != nil {
		return err
	}
	if err := scope.Provide(secretMasker); err != nil {
		return err
	}
	if err := scope.Provide(problemScanner); err != nil {
		return err
	}
	if err := scope.Provide(expressionEnv); err != nil {
		return err
	}

	if err := scope.Provide(newJobRunDecorator); err != nil {
		return err
	}
	if err := scope.Provide(newStepRunDecorator); err != nil {
		return err
	}
	if err := scope.Provide(newActionRunDecorator); err != nil {
		return err
	}

	return scope.Provide(sysEnvProvider, dig.Group(wire.EnvProvider))
}

func (w *Worker) initDossier() error {
	w.dossier = new(records.Dossier)
	if err := model.Decode(w.msg.ContextData, w.dossier); err != nil {
		return fmt.Errorf("decode ContextData: %w", err)
	}

	// GitHub context
	// https://github.com/actions/runner/blob/v2.324.0/src/Runner.Worker/ExecutionContext.cs#L882-L891
	github := w.dossier.Github
	if github.Token == "" {
		if v, ok := w.msg.Variables["system.github.token"]; ok {
			github.Token = v.Value
		} else if v, ok = w.msg.Variables["github_token"]; ok {
			github.Token = v.Value
		}
	}
	if github.Job == "" {
		if v, ok := w.msg.Variables["system.github.job"]; ok {
			github.Job = v.Value
		}
	}

	return nil
}

// https://github.com/actions/runner/blob/v2.335.1/src/Runner.Worker/ExecutionContext.cs#L1394-L1417
func (w *Worker) initFlags(scope *dig.Scope) error {
	sysVar := w.msg.Variables
	dVar := w.dossier.Variables
	if _, ok := sysVar[wire.RunnerDebug]; !ok {
		if v, ok := dVar[wire.RunnerDebug]; ok {
			sysVar[wire.RunnerDebug] = messages.Variable{Value: v}
		}
	}
	if _, ok := sysVar[wire.StepDebug]; !ok {
		if v, ok := dVar[wire.StepDebug]; ok {
			sysVar[wire.StepDebug] = messages.Variable{Value: v}
		}
	}
	var flags feature.Flags = variableFlags(sysVar)
	if err := xdig.Supply(scope, flags); err != nil {
		return fmt.Errorf("supply feature.Flags: %w", err)
	}
	return nil
}

func (w *Worker) initOtel(ctx context.Context) (context.Context, func(*error)) {
	gh := w.dossier.Github
	// TODO set LogLevel=Debug if RunnerDebug=true
	return xotel.SetupTelemetry(ctx, "worker",
		xotel.Repo(gh.Repository),
		xotel.Workflow(gh.Workflow),
		xotel.Run(gh.RunId+"#"+gh.RunAttempt),
		xotel.Job(gh.Job),
	)
}

func (w *Worker) initDiary(scope *dig.Scope) error {
	var diary scribe.Diary
	if err := xdig.Populate(scope, &diary); err != nil {
		return err
	}

	w.ctx = scribe.ContextWithScribe(w.ctx, diary)
	return nil
}

func secretMasker(req *messages.PipelineAgentJobRequest) (secret.Masker, error) {
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
				return nil, fmt.Errorf("invalid regex %q: %w", s.Value, err)
			} else {
				sm.AddSecret(secret.NewRegexSecret(re))
			}
		default:
			return nil, fmt.Errorf("unknown mask type %q", s.Type)
		}
	}
	if res := req.Resources; res != nil {
		for _, ep := range res.Endpoints {
			authz := ep.Authorization
			if authz == nil {
				continue
			}
			for _, v := range authz.Parameters {
				if v != "" {
					sm.AddSecret(secret.NewValueSecret(v))
				}
			}
		}
	}

	return sm, nil
}

func problemScanner() problem.Matchers {
	return make(problem.Matchers)
}

func expressionEnv(d *records.Dossier) (expression.Env, error) {
	opts := []expression.Option{
		expression.WithCache(true),
		expression.WithLibrary(libraries.StdLib()),
		expression.WithVariable("secrets", d.Secrets),
		expression.WithVariable("vars", d.Variables),
		expression.WithVariable("needs", d.Needs),
		expression.WithVariable("strategy", d.Strategy),
		expression.WithVariable("matrix", d.Matrix),
		expression.WithVariable("inputs", d.Inputs),
	}
	return expression.NewEnv(opts...)
}

func sysEnvProvider(req *messages.PipelineAgentJobRequest) (executor.EnvProvider, error) {
	// https://github.com/actions/runner/blob/v2.323.0/src/Runner.Worker/Handlers/NodeScriptActionHandler.cs#L53-L78
	// https://github.com/actions/runner/blob/v2.323.0/src/Runner.Worker/Handlers/ContainerActionHandler.cs#L218-L238
	sysCon := req.ServiceEndpoint("SystemVssConnection")
	if sysCon == nil {
		return nil, fmt.Errorf("service endpoint 'SystemVssConnection' not found")
	}
	var accessToken string
	if authz := sysCon.Authorization; authz != nil && authz.Scheme == "OAuth" {
		accessToken = authz.Parameters["AccessToken"]
	}
	sysEnv := map[string]string{
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

	return executor.StaticEnv(sysEnv), nil
}

func (w *Worker) newContainerRuntime(
	engine container.Engine,
	sandbox sandboxer.Sandbox,
	info *records.JobInfo,
	gh *records.Github,
) (runtime.Container, error) {
	return wire_runtime.NewContainerRuntime(w.ctx, engine, sandbox, info, gh)
}

type jobRunDecoratorParam struct {
	dig.In

	Timeline executor.JobRunDecorator `name:"timeline"`
	Log      executor.JobRunDecorator `name:"log"`
}

func newJobRunDecorator(p jobRunDecoratorParam) executor.JobRunDecorator {
	return executor.MultiJobRunDecorator{
		p.Timeline,
		p.Log,
	}
}

type stepRunDecoratorParam struct {
	dig.In

	Timeline executor.StepRunDecorator `name:"timeline"`
	Log      executor.StepRunDecorator `name:"log"`
}

func newStepRunDecorator(p stepRunDecoratorParam) executor.StepRunDecorator {
	return executor.MultiStepRunDecorator{
		p.Timeline,
		p.Log,
	}
}

type actionRunDecoratorParam struct {
	dig.In

	ConsoleCommand executor.ActionRunDecorator `name:"command"`
}

func newActionRunDecorator(p actionRunDecoratorParam) executor.ActionRunDecorator {
	return p.ConsoleCommand
}
