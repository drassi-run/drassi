package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"

	"drassi.run/core/util/context"
	"drassi.run/core/util/http"
	"drassi.run/gha-runner/pkg/messages"
	"github.com/coder/websocket"
)

const (
	taskLogEndpoint        = "%s/_apis/distributedtask/hubs/%s/plans/%s/logs"
	taskAttachmentEndpoint = "%s/_apis/distributedtask/hubs/%s/plans/%s/timelines/%s/records/%s/attachments"
)

func NewTaskService(url string, hc *http.Client, msg *messages.PipelineAgentJobRequest) (*TaskService, error) {
	client, err := newClient(url, hc)
	if err != nil {
		return nil, err
	}

	svc := &TaskService{
		client:      client,
		scopeUid:    msg.Plan.ScopeIdentifier,
		planType:    msg.Plan.PlanType,
		planUid:     msg.Plan.PlanId,
		timelineUid: msg.Timeline.Id,
	}
	return svc, nil
}

// https://github.com/actions/runner/blob/v2.323.0/src/Runner.Common/JobServer.cs
type TaskService struct {
	client      *xhttp.Client
	scopeUid    string // from jobRequest.plan.scopeIdentifier
	planType    string // from jobRequest.plan.planType (a.k.a hubName in C#)
	planUid     string // from jobRequest.plan.planId
	timelineUid string // from jobRequest.timeline.id
}

func (s *TaskService) LogUploader(recordId string) Uploader {
	return &logTaskUploader{svc: s, recordId: recordId}
}

// https://github.com/actions/runner/blob/v2.323.0/src/Runner.Common/JobServerQueue.cs#L882-L896
func (s *TaskService) uploadLog(ctx context.Context, recordId string, r io.Reader) error {
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

func (s *TaskService) AttachmentUploader(recordId, kind, name string) Uploader {
	return &attachmentTaskUploader{svc: s, recordId: recordId, kind: kind, name: name}
}

// https://github.com/actions/runner/blob/v2.323.0/src/Runner.Common/JobServerQueue.cs#L900-L903
func (s *TaskService) uploadAttach(ctx context.Context, recordId, kind, name string, r io.Reader) error {
	endpoint := fmt.Sprintf(taskAttachmentEndpoint, s.scopeUid, s.planType, s.planUid, s.timelineUid, recordId)
	endpoint = path.Join(endpoint, kind, name)

	e := s.client.Post(endpoint).
		SetQuery("api-version", "5.1-preview").
		SetHeader("Content-Type", "application/octet-stream").
		WithBody(r)

	return e.Do(ctx)
}

func (s *TaskService) LiveFeeder(contextual xcontext.Provider, wsUrl string) (LiveFeeder, error) {
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

	lf := &taskLiveFeeder{
		svc:    s,
		wsConn: wsConn,
	}
	return lf, nil
}

func (s *TaskService) RecordTimeline(event any) error {
	return nil
}

type logTaskUploader struct {
	svc      *TaskService
	recordId string
}

func (u *logTaskUploader) Upload(ctx context.Context, r io.Reader) error {
	return u.svc.uploadLog(ctx, u.recordId, r)
}

func (u *logTaskUploader) Complete(context.Context, int64) error {
	return nil
}

type attachmentTaskUploader struct {
	svc                  *TaskService
	recordId, kind, name string
}

func (u *attachmentTaskUploader) Upload(ctx context.Context, r io.Reader) error {
	return u.svc.uploadAttach(ctx, u.recordId, u.kind, u.name, r)
}

func (u *attachmentTaskUploader) Complete(context.Context, int64) error {
	return nil
}

type taskLiveFeeder struct {
	svc    *TaskService
	wsConn *websocket.Conn
}

// First, try push the log via websocket if avaiable, otherwise fallback to the REST API.
//
// https://github.com/actions/runner/blob/v2.324.0/src/Runner.Common/JobServerQueue.cs#L418
// https://github.com/actions/runner/blob/v2.324.0/src/Runner.Common/JobServer.cs#L229
// https://github.com/actions/runner/blob/v2.324.0/src/Sdk/DTGenerated/Generated/TaskHttpClientBase.cs#L115
func (lf *taskLiveFeeder) Handle(ctx context.Context, s string) error {
	//TODO implement me
	panic("implement me")
}

func (lf *taskLiveFeeder) Start() error {
	//TODO implement me
	panic("implement me")
}

func (lf *taskLiveFeeder) Close() error {
	//TODO implement me
	panic("implement me")
}
