package service

import (
	"context"
	"net/http"
	"time"

	"drassi.run/core/util/http"
	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/types"
	"github.com/chainguard-dev/clog"
)

func newClient(url string, hc *http.Client) (*xhttp.Client, error) {
	client, err := xhttp.NewClient(url)
	if err != nil {
		return nil, err
	}

	client = client.WithDefaultErrorHandler(types.ParseActionsError).
		WithDefaultHeader("User-Agent", "gha-runner") // TODO

	if hc != nil {
		client = client.WithHttpClient(hc)
	}
	return client, nil
}

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

func (s *RunService) RenewJob(ctx context.Context, msg *messages.PipelineAgentJobRequest) {
	l := clog.FromContext(ctx)
	req := &renewJobRequest{
		PlanID: msg.Plan.PlanId,
		JobId:  msg.JobId,
	}
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
		l.DebugContextf(ctx, "successfully renew job %s, job is valid till %s", msg.JobId, resp.LockedUntil)
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

func (s *RunService) CompleteJob(ctx context.Context, msg *messages.PipelineAgentJobRequest) error {
	req := &completeJobRequest{
		PlanID: msg.Plan.PlanId,
		JobID:  msg.JobId,
		// TODO
		//Conclusion:     conclusion,
		//Outputs:        outputs,
		//StepResults:    stepResults,
		//Annotations:    jobAnnotations,
		//EnvironmentUrl: environmentUrl,
		//Telemetry:      telemetry,
		//BillingOwnerId: billingOwnerId,
	}
	hr := s.client.Post("completejob").
		WithBodyProvider(xhttp.JsonEncode(req))

	return hr.Do(ctx)
}
