/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

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
	"drassi.run/core/pkg/executor/command/cmdtypes"
	"drassi.run/core/util/http"
	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/timeline"
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
		OnSuccess(xhttp.JsonDecode(msg, messages.JsonOptions()...))

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
	timer := time.NewTimer(0)
	defer timer.Stop()

	doRenew := func() error {
		resp := new(renewJobResponse)
		hr := s.client.Post("renewjob").
			WithBodyProvider(xhttp.JsonEncode(req)).
			OnSuccess(xhttp.JsonDecode(resp, messages.JsonOptions()...))

		if err := hr.Do(ctx); err != nil {
			return err
		}
		l.Debugf("successfully renew job %s, job is valid till %s", req.JobId, resp.LockedUntil)
		if d := renewAt(resp.LockedUntil); d >= 0 {
			timer.Reset(d)
		}
		return nil
	}

	for {
		select {
		case <-timer.C:
			if err := doRenew(); err != nil {
				l.Errorf("renew job failed: %v", err)
				return
			}
		case <-ctx.Done():
			l.Debugf("renew job completed")
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
		PlanId: l.msg.Plan.PlanId,
		JobId:  l.msg.JobId,
	}
}

func (l *runLease) Complete(ctx context.Context, record *timeline.Record) error {
	if l.done != nil {
		l.done() // cancel Renew
	}

	if req, err := l.completeRequest(record); err != nil {
		return err
	} else {
		return l.svc.completeJob(ctx, req)
	}
}

func (l *runLease) completeRequest(r *timeline.Record) (*completeJobRequest, error) {
	job, ok := r.Object.(*timeline.JobObject)
	if !ok {
		return nil, fmt.Errorf("%T is not *JobObject", r.Object)
	}

	stepResults, err := l.toStepResults(r.Children)
	if err != nil {
		return nil, err
	}

	req := &completeJobRequest{
		PlanId:         l.msg.Plan.PlanId,
		JobId:          l.msg.JobId,
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

func (l *runLease) toAnnotations(issues []*cmdtypes.Issue) []*Annotation {
	annotations := make([]*Annotation, 0, len(issues))
	for _, iss := range issues {
		if anno := l.toAnnotation(iss); anno != nil {
			annotations = append(annotations, anno)
		}
	}
	return annotations
}

// https://github.com/actions/runner/blob/v2.324.0/src/Sdk/RSWebApi/Contracts/IssueExtensions.cs#L7
func (l *runLease) toAnnotation(iss *cmdtypes.Issue) *Annotation {
	var msg string
	if m := iss.Message; m != "" {
		msg = m
	} else {
		msg = iss.Data["message"]
	}
	if msg = strings.TrimSpace(msg); msg == "" {
		return nil
	}

	a := &Annotation{
		Message: msg,
		Level:   ToAnnotationLevel(iss.Type),
	}

	if file := iss.Data["file"]; file != "" {
		a.Path = file
	}
	if title := iss.Data["title"]; title != "" {
		a.Title = title
	}
	if s := iss.Data["line"]; s != "" {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			a.StartLine = i
		}
	}
	if s := iss.Data["endLine"]; s != "" {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			a.EndLine = i
		}
	}
	if s := iss.Data["col"]; s != "" {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			a.StartColumn = i
		}
	}
	if s := iss.Data["endColumn"]; s != "" {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			a.EndColumn = i
		}
	}
	if s := iss.Data["stepNumber"]; s != "" {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			a.StepNumber = i
		}
	}
	if s := iss.Data["logFileLineNumber"]; s != "" {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil && i != 0 {
			if a.Path == "" && a.StartLine == 0 {
				a.StartLine = i
				a.EndLine = i
			}
		}
	}
	return a
}

func (l *runLease) toStepResults(records []*timeline.Record) ([]*StepResult, error) {
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

func (l *runLease) toStepResult(r *timeline.Record) (*StepResult, error) {
	step, ok := r.Object.(*timeline.StepObject)
	if !ok {
		return nil, fmt.Errorf("%T is not a *StepObject", r.Object)
	}

	res := &StepResult{
		Id:          r.Uid,
		Number:      r.Order,
		Name:        r.Name,
		Status:      r.State,
		Conclusion:  r.Result,
		StartedAt:   r.StartedAt,
		CompletedAt: r.CompletedAt,
		Annotations: l.toAnnotations(r.Issues),
	}

	// https://github.com/actions/runner/blob/v2.324.0/src/Runner.Worker/Handlers/Handler.cs#L57
	switch action := step.StepSpec.Action.(type) {
	case *executor.ScriptActionSpec:
		res.ActionType = "run"
	case *executor.DockerActionSpec:
		res.ActionType = "docker"
	case *executor.ReferenceActionSpec:
		res.ActionType = "repository"

		repo := action.Repository()
		res.ActionRef = repo.Ref
		if repo.Path == "" {
			res.ActionName = repo.Name
		} else {
			res.ActionName = path.Join(repo.Name, repo.Path)
		}
	}

	return res, nil
}
