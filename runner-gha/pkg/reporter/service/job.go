package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"

	"drassi.run/core/util/context"
	"drassi.run/core/util/http"
	"drassi.run/gha-runner/pkg/holder"
	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/types"
	"github.com/coder/websocket"
)

const (
	taskLogEndpoint        = "%s/_apis/distributedtask/hubs/%s/plans/%s/logs"
	taskAttachmentEndpoint = "%s/_apis/distributedtask/hubs/%s/plans/%s/timelines/%s/records/%s/attachments"
)

func NewJobService(url string, hc *http.Client, msg *messages.PipelineAgentJobRequest) (*JobService, error) {
	client, err := newClient(url, hc)
	if err != nil {
		return nil, err
	}

	svc := &JobService{
		client:      client,
		scopeUid:    msg.Plan.ScopeIdentifier,
		planType:    msg.Plan.PlanType,
		planUid:     msg.Plan.PlanId,
		timelineUid: msg.Timeline.Id,
	}
	return svc, nil
}

// https://github.com/actions/runner/blob/v2.323.0/src/Runner.Common/JobServer.cs
type JobService struct {
	client      *xhttp.Client
	scopeUid    string // from jobRequest.plan.scopeIdentifier
	planType    string // from jobRequest.plan.planType (a.k.a hubName in C#)
	planUid     string // from jobRequest.plan.planId
	timelineUid string // from jobRequest.timeline.id
}

func (s *JobService) LogUploader(recordId string) Uploader {
	return &logJobUploader{svc: s, recordId: recordId}
}

// https://github.com/actions/runner/blob/v2.323.0/src/Runner.Common/JobServerQueue.cs#L882-L896
func (s *JobService) uploadLog(ctx context.Context, recordId string, r io.Reader) error {
	// Create the log
	tl := new(taskLog)
	endpoint := fmt.Sprintf(taskLogEndpoint, s.scopeUid, s.planType, s.planUid)
	e := s.client.Post(endpoint).
		SetQuery("api-version", "5.1-preview").
		WithBodyProvider(xhttp.JsonEncode(&taskLog{Path: `log\` + recordId})).
		OnSuccess(xhttp.JsonDecode(tl))

	if err := e.Do(ctx); err != nil {
		return err
	}

	// Upload the contents
	endpoint = path.Join(endpoint, tl.Id)
	e = s.client.Post(endpoint).
		SetQuery("api-version", "5.1-preview").
		SetHeader("Content-Type", "application/octet-stream").
		WithBody(r)

	return e.Do(ctx)

	// TODO: Create a new record and only set the Log field
}

func (s *JobService) AttachmentUploader(recordId, kind, name string) Uploader {
	return &attachmentJobUploader{svc: s, recordId: recordId, kind: kind, name: name}
}

// https://github.com/actions/runner/blob/v2.323.0/src/Runner.Common/JobServerQueue.cs#L900-L903
func (s *JobService) uploadAttach(ctx context.Context, recordId, kind, name string, r io.Reader) error {
	endpoint := fmt.Sprintf(taskAttachmentEndpoint, s.scopeUid, s.planType, s.planUid, s.timelineUid, recordId)
	endpoint = path.Join(endpoint, kind, name)

	e := s.client.Post(endpoint).
		SetQuery("api-version", "5.1-preview").
		SetHeader("Content-Type", "application/octet-stream").
		WithBody(r)

	return e.Do(ctx)
}

func (s *JobService) LiveFeeder(contextual xcontext.Provider, wsUrl string) (LiveFeeder, error) {
	ctx := contextual.Context()
	opts := &websocket.DialOptions{
		HTTPClient: s.client.HttpClient(),
	}

	var wsConn *websocket.Conn
	if wsUrl != "" {
		if conn, _, err := websocket.Dial(ctx, wsUrl, opts); err != nil {
			return nil, err
		} else {
			wsConn = conn
		}
	}

	lf := &jobLiveFeeder{
		svc:    s,
		wsConn: wsConn,
	}
	return lf, nil
}

func (s *JobService) RecordTimeline(event any) error {
	return nil
}

type logJobUploader struct {
	svc      *JobService
	recordId string
}

func (u *logJobUploader) Upload(ctx context.Context, r io.Reader) error {
	return u.svc.uploadLog(ctx, u.recordId, r)
}

func (u *logJobUploader) Complete(context.Context, int64) error {
	return nil
}

type attachmentJobUploader struct {
	svc                  *JobService
	recordId, kind, name string
}

func (u *attachmentJobUploader) Upload(ctx context.Context, r io.Reader) error {
	return u.svc.uploadAttach(ctx, u.recordId, u.kind, u.name, r)
}

func (u *attachmentJobUploader) Complete(context.Context, int64) error {
	return nil
}

type jobLiveFeeder struct {
	svc    *JobService
	wsConn *websocket.Conn
}

// First, try push the log via websocket if avaiable, otherwise fallback to the REST API.
//
// https://github.com/actions/runner/blob/v2.324.0/src/Runner.Common/JobServerQueue.cs#L418
// https://github.com/actions/runner/blob/v2.324.0/src/Runner.Common/JobServer.cs#L229
// https://github.com/actions/runner/blob/v2.324.0/src/Sdk/DTGenerated/Generated/TaskHttpClientBase.cs#L115
func (lf *jobLiveFeeder) Handle(ctx context.Context, s string) error {
	//TODO implement me
	panic("implement me")
}

func (lf *jobLiveFeeder) Start() error {
	//TODO implement me
	panic("implement me")
}

func (lf *jobLiveFeeder) Close() error {
	//TODO implement me
	panic("implement me")
}

func (s *JobService) WrapLease(l holder.Lease) holder.Lease {
	return l
}

func (s *JobService) completeJob(ctx context.Context, record *types.Record) error {
	return nil // TODO
}

type jobLeaseWrapper struct {
	holder.Lease
	svc *JobService
}

func (l *jobLeaseWrapper) Complete(ctx context.Context, record *types.Record) error {
	if err := l.svc.completeJob(ctx, record); err != nil {
		return err
	}
	return l.Lease.Complete(ctx, record)
}
