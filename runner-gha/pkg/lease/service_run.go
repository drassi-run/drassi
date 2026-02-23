package lease

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/support"
	"drassi.run/core/util/http"
	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/types"
	"github.com/chainguard-dev/clog"
)

func NewRunService(url string, hc *http.Client) (*RunService, error) {
	client, err := newClient(url, hc)
	if err != nil {
		return nil, err
	}

	s := &RunService{
		client: client,
	}
	return s, nil
}

type RunService struct {
	client *xhttp.Client
}

func (s *RunService) AcquireJob(ctx context.Context, msgId, billingOwner string) (*messages.PipelineAgentJobRequest, error) {
	o := acquireJobRequest{
		JobMessageId:   msgId,
		RunnerOS:       "Linux",
		BillingOwnerId: billingOwner,
	}
	msg := new(messages.PipelineAgentJobRequest)
	hr := s.client.Post("acquirejob").
		WithBodyProvider(xhttp.JsonEncode(&o)).
		OnSuccess(xhttp.JsonDecode(msg))

	if err := hr.Do(ctx); err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *RunService) Lease(msg *messages.PipelineAgentJobRequest) Lease {
	return &runLease{svc: s, msg: msg}
}

func (s *RunService) renewJob(ctx context.Context, req *renewJobRequest) {
	l := clog.FromContext(ctx)
	resp := new(renewJobResponse)
	hr := s.client.Post("renewjob").
		WithBodyProvider(xhttp.JsonEncode(req)).
		OnSuccess(xhttp.JsonDecode(resp))
	timer := time.NewTimer(0)
	defer timer.Stop()

	doRenew := func() {
		if err := hr.Do(ctx); err != nil {
			l.ErrorContextf(ctx, "renewjob failed: %v", err)
			return
		}
		l.DebugContextf(ctx, "successfully renew job %s, job is valid till %s", req.JobId, resp.LockedUntil)
		if d := renewAt(resp.LockedUntil); d >= 0 {
			timer.Reset(d)
		}
	}

	for {
		select {
		case <-timer.C:
			doRenew()
		case <-ctx.Done():
			return
		}
	}
}

func (s *RunService) completeJob(ctx context.Context, req *completeJobRequest) error {
	hr := s.client.Post("completejob").
		WithBodyProvider(xhttp.JsonEncode(req))

	return hr.Do(ctx)
}

type runLease struct {
	svc  *RunService
	msg  *messages.PipelineAgentJobRequest
	done context.CancelFunc
}

func (l *runLease) GetMessage() *messages.PipelineAgentJobRequest {
	return l.msg
}

func (l *runLease) Renew(ctx context.Context) {
	ctx, l.done = context.WithCancel(ctx)

	req := l.renewRequest()
	l.svc.renewJob(ctx, req)
}

func (l *runLease) renewRequest() *renewJobRequest {
	return &renewJobRequest{
		PlanID: l.msg.Plan.PlanId,
		JobId:  l.msg.JobId,
	}
}

func (l *runLease) Complete(ctx context.Context, record *types.Record) error {
	l.done() // cancel Renew

	if req, err := l.completeRequest(record); err != nil {
		return err
	} else {
		return l.svc.completeJob(ctx, req)
	}
}

func (l *runLease) completeRequest(r *types.Record) (*completeJobRequest, error) {
	job, ok := r.Object.(*types.JobObject)
	if !ok {
		return nil, fmt.Errorf("%T is not *JobObject", r.Object)
	}

	stepResults, err := l.toStepResults(r.Children)
	if err != nil {
		return nil, err
	}

	req := &completeJobRequest{
		PlanID:         l.msg.Plan.PlanId,
		JobID:          l.msg.JobId,
		BillingOwnerId: l.msg.BillingOwnerId,

		Conclusion:     r.Result,
		Outputs:        l.convertOutputs(job.Outputs),
		EnvironmentUrl: job.EnvironmentUrl,

		StepResults: stepResults,
		Annotations: l.toAnnotations(r.Issues),
	}

	return req, nil
}

func (l *runLease) convertOutputs(m map[string]string) map[string]messages.Variable {
	res := make(map[string]messages.Variable, len(m))
	for k, v := range m {
		res[k] = messages.Variable{
			Value:    v,
			IsSecret: false,
		}
	}

	return res
}

func (l *runLease) toAnnotations(issues []*support.Issue) []*Annotation {
	annotations := make([]*Annotation, 0, len(issues))
	for _, issue := range issues {
		if anno := l.toAnnotation(issue); anno != nil {
			annotations = append(annotations, anno)
		}
	}
	return annotations
}

// https://github.com/actions/runner/blob/v2.324.0/src/Sdk/RSWebApi/Contracts/IssueExtensions.cs#L7
func (l *runLease) toAnnotation(issue *support.Issue) *Annotation {
	var msg string
	if m := issue.Message; m != "" {
		msg = m
	} else {
		msg = issue.Data["message"]
	}
	if msg = strings.TrimSpace(msg); msg == "" {
		return nil
	}

	a := &Annotation{
		Message: msg,
		Level:   ToAnnotationLevel(issue.Type),
	}

	if file := issue.Data["file"]; file != "" {
		a.Path = file
	}
	if title := issue.Data["title"]; title != "" {
		a.Title = title
	}
	if s := issue.Data["line"]; s != "" {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			a.StartLine = i
		}
	}
	if s := issue.Data["endLine"]; s != "" {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			a.EndLine = i
		}
	}
	if s := issue.Data["col"]; s != "" {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			a.StartColumn = i
		}
	}
	if s := issue.Data["endColumn"]; s != "" {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			a.EndColumn = i
		}
	}
	if s := issue.Data["stepNumber"]; s != "" {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			a.StepNumber = i
		}
	}
	if s := issue.Data["logFileLineNumber"]; s != "" {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil && i != 0 {
			if a.Path == "" && a.StartLine == 0 {
				a.StartLine = i
				a.EndLine = i
			}
		}
	}
	return a
}

func (l *runLease) toStepResults(records []*types.Record) ([]*StepResult, error) {
	results := make([]*StepResult, 0, len(records))

	for _, r := range records {
		if res, err := l.toStepResult(r); err != nil {
			return nil, err
		} else {
			results = append(results, res)
		}
	}

	return results, nil
}

func (l *runLease) toStepResult(r *types.Record) (*StepResult, error) {
	step, ok := r.Object.(*types.StepObject)
	if !ok {
		return nil, fmt.Errorf("%T is not a *StepObject", r.Object)
	}

	res := &StepResult{
		Id:          r.Uid,
		Number:      r.Order,
		Name:        step.StepRun.DisplayName(step.Stage),
		Status:      r.State,
		Conclusion:  r.Result,
		StartedAt:   r.StartedAt,
		CompletedAt: r.CompletedAt,
		Annotations: l.toAnnotations(r.Issues),
	}

	// https://github.com/actions/runner/blob/v2.324.0/src/Runner.Worker/Handlers/Handler.cs#L57
	switch sr := step.StepRun.(type) {
	case *executor.ScriptStepRun:
		res.ActionType = "run"
	case *executor.DockerStepRun:
		res.ActionType = "docker"
	case *executor.ActionStepRun:
		res.ActionType = "repository"

		repo := sr.Repository()
		res.ActionRef = repo.Ref
		if repo.Path == "" {
			res.ActionName = repo.Name
		} else {
			res.ActionName = path.Join(repo.Name, repo.Path)
		}
	}

	return res, nil
}
