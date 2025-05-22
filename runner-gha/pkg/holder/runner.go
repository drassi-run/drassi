package holder

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"drassi.run/core/util/http"
	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/types"
	"github.com/chainguard-dev/clog"
)

func NewRunnerService(url string, hc *http.Client, groupId int) (*RunnerService, error) {
	client, err := newClient(url, hc)
	if err != nil {
		return nil, err
	}

	s := &RunnerService{
		client:  client,
		groupId: groupId,
	}
	return s, nil
}

// https://github.com/actions/runner/blob/v2.323.0/src/Runner.Listener/JobDispatcher.cs#L383
var lockToken = "00000000-0000-0000-0000-000000000000" // Guid.Empty

// RunnerService implements
//   - the job request methods of [RunnerServer]
//   - [ActionsRunServer]
//
// of the C# GitHub actions/runner
//
// [RunnerServer]: https://github.com/actions/runner/blob/v2.324.0/src/Runner.Common/RunnerServer.cs#L44-L46
// [ActionsRunServer]: https://github.com/actions/runner/blob/v2.324.0/src/Runner.Common/ActionsRunServer.cs#L16
type RunnerService struct {
	client  *xhttp.Client
	groupId int
}

// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/DTWebApi/WebApi/ActionsRunServerHttpClient.cs#L71
func (s *RunnerService) AcquireJob(ctx context.Context, messageId string) (*messages.PipelineAgentJobRequest, error) {
	msg := new(messages.PipelineAgentJobRequest)
	hr := s.client.Get(fmt.Sprintf("_apis/distributedtask/runnermessages/%s", messageId)).
		SetQuery("api-version", "6.0-preview").
		OnSuccess(xhttp.JsonDecode(msg))

	if err := hr.Do(ctx); err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *RunnerService) Lease(msg *messages.PipelineAgentJobRequest) Lease {
	return &runnerLease{svc: s, msg: msg}
}

// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/DTWebApi/WebApi/TaskAgentHttpClient.cs#L93
func (s *RunnerService) renewJob(ctx context.Context, msg *messages.PipelineAgentJobRequest, orchId string) {
	l := clog.FromContext(ctx)
	req := &runnerJobRequest{
		RequestId:   msg.RequestId,
		LockedUntil: msg.LockedUntil.Time,
	}
	resp := new(runnerJobRequest)
	hr := s.client.Patch(fmt.Sprintf("_apis/distributedtask/pools/%d/jobrequests/%d", s.groupId, msg.RequestId)).
		SetQuery("api-version", "5.1-preview").
		SetQuery("lockToken", lockToken).
		WithBodyProvider(xhttp.JsonEncode(req)).
		OnSuccess(xhttp.JsonDecode(resp))

	if orchId != "" {
		hr.SetHeader("X-VSS-OrchestrationId", orchId)
	}

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

// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/DTWebApi/WebApi/TaskAgentHttpClient.cs#L61
func (s *RunnerService) completeJob(ctx context.Context, msg *messages.PipelineAgentJobRequest, result types.Result) error {
	req := &runnerJobRequest{
		RequestId:  msg.RequestId,
		FinishTime: time.Now(),
		Result:     result,
	}

	hr := s.client.Patch(fmt.Sprintf("_apis/distributedtask/pools/%d/jobrequests/%d", s.groupId, msg.RequestId)).
		SetQuery("api-version", "5.1-preview").
		SetQuery("lockToken", lockToken).
		WithBodyProvider(xhttp.JsonEncode(req))

	return hr.Do(ctx)
}

type runnerLease struct {
	svc *RunnerService
	msg *messages.PipelineAgentJobRequest
}

func (l *runnerLease) GetMessage() *messages.PipelineAgentJobRequest {
	return l.msg
}

func (l *runnerLease) Renew(ctx context.Context) {
	orchId := ""
	// orchId also can be extracted from `orch_id` claim
	// in JWT token from msg.Resources.Endpoints.Authorization.Parameters["AccessToken"]
	if v, ok := l.msg.Variables["system.orchestrationId"]; ok {
		orchId = v.Value
	}
	l.svc.renewJob(ctx, l.msg, orchId)
}

func (l *runnerLease) Complete(ctx context.Context, record *types.Record) error {
	return l.svc.completeJob(ctx, l.msg, record.Result)
}
