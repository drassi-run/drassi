package executor

import (
	"context"
	"fmt"
	"strings"

	"drassi.run/core/pkg/executor/logging"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/util/dig"
	"drassi.run/core/util/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/dig"
)

type telemetryStepRun struct {
	StepRun
	supervisor Supervisor
}

func WithTelemetryStepRun(sr StepRun) StepRun {
	if _, ok := sr.(*telemetryStepRun); ok {
		return sr
	}

	if csr, ok := sr.(*CompositeStepRun); ok {
		for i, s := range csr.StepRuns {
			csr.StepRuns[i] = WithTelemetryStepRun(s)
		}
	}
	return &telemetryStepRun{StepRun: sr}
}

func (sr *telemetryStepRun) Initialize(ctx context.Context, scope *dig.Scope) (err error) {
	if err = xdig.Populate(scope, &sr.supervisor); err != nil {
		return err
	}

	ctx, span := xotel.StartSpan(ctx, fmt.Sprintf("Task.Initialize(%s)", sr.StepId()))
	defer xotel.EndSpan(span, &err)

	stop := sr.supervisor.StartContext(ctx)
	defer stop()

	return sr.StepRun.Initialize(ctx, scope)
}

func (sr *telemetryStepRun) PreTask() *Task {
	return sr.decorateTask(StepRun.PreTask)
}

func (sr *telemetryStepRun) MainTask() *Task {
	return sr.decorateTask(StepRun.MainTask)
}

func (sr *telemetryStepRun) PostTask() *Task {
	return sr.decorateTask(StepRun.PostTask)
}

func (sr *telemetryStepRun) decorateTask(fn func(StepRun) *Task) *Task {
	task := fn(sr.StepRun)
	if task == nil {
		return nil
	}

	delegate := task.Run
	task.Run = func(ctx context.Context, se StepExecutor) (err error) {
		spanName := fmt.Sprintf("Task.Run(%s)", sr.StepId())
		ctx, span := xotel.StartSpan(ctx, spanName, trace.WithAttributes(
			xotel.DrassiStage(string(task.Stage)),
			xotel.DrassiStep(task.StepId),
		))
		defer xotel.EndSpan(span, &err)

		stop := sr.supervisor.StartContext(ctx)
		defer stop()

		return delegate(ctx, se)
	}
	return task
}

type telemetryStepExecutor struct {
	StepExecutor
	supervisor Supervisor
}

func WithTelemetryStepExecutor(exec StepExecutor) StepExecutor {
	if _, ok := exec.(*telemetryStepExecutor); ok {
		return exec
	}
	return &telemetryStepExecutor{StepExecutor: exec}
}

func (e *telemetryStepExecutor) Initialize(ctx context.Context, scope *dig.Scope) (err error) {
	stepId := StepId(e)
	ctx, span := xotel.StartSpan(ctx, fmt.Sprintf("StepExecutor.Initialize(%s)", stepId),
		trace.WithAttributes(xotel.DrassiStep(FullStepId(e))),
	)
	defer xotel.EndSpan(span, &err)

	ctx, syslog := xotel.ChildLogger(ctx, "step", stepId)
	syslog.Infof("initialize step %q", stepId)

	if err = xdig.Populate(scope, &e.supervisor); err != nil {
		syslog.Errorf("failed to populate supervisor: %v", err)
		return err
	}

	stop := e.supervisor.StartContext(ctx)
	defer stop()

	return e.StepExecutor.Initialize(ctx, scope)
}

func (e *telemetryStepExecutor) RunStep(ctx context.Context, fn func(StepRun) *Task) *records.Step {
	task := fn(e.StepRun())
	if task == nil {
		return nil
	}

	stepId := StepId(e)
	ctx, span := xotel.StartSpan(ctx, fmt.Sprintf("StepExecutor.RunStep(%s)", stepId),
		trace.WithAttributes(
			xotel.DrassiStage(string(task.Stage)),
			xotel.DrassiStep(FullStepId(e)),
		),
	)
	defer span.End()

	ctx, syslog := xotel.ChildLogger(ctx, "step", stepId)
	syslog.Infof("running step %q", stepId)

	stop := e.supervisor.StartContext(ctx)
	defer stop()

	res := e.StepExecutor.RunStep(ctx, fn)
	syslog.Infof("%s step %q completed with outcome=%q conclusion=%q", task.Stage, stepId, res.Outcome, res.Conclusion)
	return res
}

type telemetryJobExecutor struct {
	JobExecutor
	supervisor Supervisor
}

func WithTelemetryJobExecutor(exec JobExecutor) JobExecutor {
	if _, ok := exec.(*telemetryJobExecutor); ok {
		return exec
	}
	return &telemetryJobExecutor{JobExecutor: exec}
}

func (e *telemetryJobExecutor) Initialize(ctx context.Context, scope *dig.Scope) (err error) {
	jobId := JobId(e)
	ctx, span := xotel.StartSpan(ctx, fmt.Sprintf("JobExecutor.Initialize(%s)", jobId),
		trace.WithAttributes(xotel.DrassiJob(jobId)),
	)
	defer xotel.EndSpan(span, &err)

	ctx, syslog := xotel.ChildLogger(ctx, "job", jobId)
	syslog.Infof("initialize job %q", jobId)

	if err = xdig.Populate(scope, &e.supervisor); err != nil {
		syslog.Errorf("failed to populate supervisor: %v", err)
		return err
	}

	stop := e.supervisor.StartContext(ctx)
	defer stop()

	return e.JobExecutor.Initialize(ctx, scope)
}

func (e *telemetryJobExecutor) RunJob(ctx context.Context) *records.Job {
	jobId := JobId(e)
	ctx, span := xotel.StartSpan(ctx, fmt.Sprintf("JobExecutor.RunJob(%s)", jobId),
		trace.WithAttributes(xotel.DrassiJob(jobId)),
	)
	defer span.End()

	ctx, syslog := xotel.ChildLogger(ctx, "job", jobId)
	syslog.Infof("running job %q", jobId)

	stop := e.supervisor.StartContext(ctx)
	defer stop()

	res := e.JobExecutor.RunJob(ctx)
	syslog.Infof("job %q completed with result=%q", jobId, res.Result)
	return res
}

func (e *telemetryJobExecutor) Finalize(ctx context.Context) (err error) {
	jobId := JobId(e)
	ctx, span := xotel.StartSpan(ctx, fmt.Sprintf("JobExecutor.Finalize(%s)", jobId),
		trace.WithAttributes(xotel.DrassiJob(jobId)),
	)
	defer xotel.EndSpan(span, &err)

	ctx, syslog := xotel.ChildLogger(ctx, "job", jobId)
	syslog.Infof("terminate job %q", jobId)

	stop := e.supervisor.StartContext(ctx)
	defer stop()

	return e.JobExecutor.Finalize(ctx)
}

func withArray(name string, a []string) func(logger logging.Logger) {
	return func(logger logging.Logger) {
		switch l := len(a); {
		case l == 0: // does nothing
		case l <= 3:
			logging.Logf(logger, "%s: [%s]", name, strings.Join(a, ", "))
		default:
			logging.Logf(logger, "%s:", name)
			for _, e := range a {
				logging.Logf(logger, "  - %s", e)
			}
		}
	}
}

func withMap(name string, m map[string]string) func(logger logging.Logger) {
	return func(logger logging.Logger) {
		if len(m) == 0 {
			return
		}
		logging.Logf(logger, "%s:", name)
		for k, v := range m {
			logging.Logf(logger, "  %s: %s", k, v)
		}
	}
}

func withKV(key string, value string) func(logger logging.Logger) {
	return func(logger logging.Logger) {
		if value == "" {
			return
		}
		logging.Logf(logger, "%s: %s", key, value)
	}
}
