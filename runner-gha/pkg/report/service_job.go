package report

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"

	"drassi.run/core/pkg/executor/support"
	"drassi.run/core/util/context"
	"drassi.run/core/util/http"
	"drassi.run/gha-runner/pkg/lease"
	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/types"
	"github.com/coder/websocket"
)

const (
	taskLogEndpoint        = "%s/_apis/distributedtask/hubs/%s/plans/%s/logs"
	taskAttachmentEndpoint = "%s/_apis/distributedtask/hubs/%s/plans/%s/timelines/%s/records/%s/attachments"
	taskTimelineEndpoint   = "%s/_apis/distributedtask/hubs/%s/plans/%s/timelines/%s/records"
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

////////////// Logs //////////////

func (s *JobService) LogUploader(recordId string) Uploader {
	return &jobLogUploader{svc: s, recordId: recordId}
}

type jobLogUploader struct {
	svc      *JobService
	recordId string
}

func (u *jobLogUploader) Upload(ctx context.Context, r io.Reader, _ *Stat) error {
	return u.svc.uploadLog(ctx, u.recordId, r)
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

////////////// Attachment //////////////

func (s *JobService) AttachmentUploader(recordId, kind, name string) Uploader {
	return &jobAttachmentUploader{svc: s, recordId: recordId, kind: kind, name: name}
}

type jobAttachmentUploader struct {
	svc                  *JobService
	recordId, kind, name string
}

func (u *jobAttachmentUploader) Upload(ctx context.Context, r io.Reader, _ *Stat) error {
	return u.svc.uploadAttach(ctx, u.recordId, u.kind, u.name, r)
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
		HTTPClient:      s.client.HttpClient(),
		CompressionMode: websocket.CompressionContextTakeover,
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

func (s *JobService) TimelineRecorder() TimelineRecorder {
	return &jobTimelineRecorder{svc: s}
}

type jobTimelineRecorder struct {
	svc *JobService
}

func (r *jobTimelineRecorder) Update(ctx context.Context, records ...*types.Record) error {
	if len(records) == 0 {
		return nil
	}

	timelineRecords := make([]*record, len(records))
	for i, rec := range records {
		timelineRecords[i] = toTimelineRecord("", r.svc.timelineUid, rec)
	}

	return r.svc.updateTimelineRecord(ctx, timelineRecords)
}

func (s *JobService) updateTimelineRecord(ctx context.Context, records []*record) error {
	endpoint := fmt.Sprintf(taskTimelineEndpoint, s.scopeUid, s.planType, s.planUid, s.timelineUid)
	body := types.NewList(records)

	e := s.client.Patch(endpoint).
		SetQuery("api-version", "5.1-preview").
		SetHeader("Content-Type", "application/octet-stream").
		WithBodyProvider(xhttp.JsonEncode(body))

	return e.Do(ctx)
}

func toTimelineRecord(parentId, timelineUid string, rec *types.Record) *record {
	r := &record{
		Id:         rec.Uid,
		ParentId:   parentId,
		Order:      rec.Order,
		TimelineId: timelineUid,
		StartTime:  rec.StartedAt,
		FinishTime: rec.CompletedAt,
		State:      rec.State,
		Result:     rec.Result,
		Issues:     rec.Issues,
	}

	for _, issue := range rec.Issues {
		switch issue.Type {
		case support.IssueTypeError:
			r.ErrorCount++
		case support.IssueTypeWarning:
			r.WarningCount++
		case support.IssueTypeNotice:
			r.NoticeCount++
		}
	}

	if rec.State == types.StateCompleted {
		r.PercentComplete = 100
	}

	// https://github.com/actions/runner/blob/v2.324.0/src/Runner.Worker/ExecutionContext.cs#L27-L31
	switch rec.Object.(type) {
	case *types.JobObject:
		r.Type = "Job"
	case *types.StepObject:
		r.Type = "Task"
	}
	return r
}

func (s *JobService) WrapLease(l lease.Lease) lease.Lease {
	return &jobLeaseWrapper{
		Lease: l,
		svc:   s,
	}
}

type jobLeaseWrapper struct {
	lease.Lease
	svc *JobService
}

func (l *jobLeaseWrapper) Complete(ctx context.Context, record *types.Record) error {
	if err := l.svc.completeJob(ctx, record); err != nil {
		return err
	}
	return l.Lease.Complete(ctx, record)
}

func (s *JobService) completeJob(ctx context.Context, record *types.Record) error {
	return nil // TODO
}
