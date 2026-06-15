/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"

	"drassi.run/core/pkg/executor/command/cmdtypes"
	"drassi.run/core/util/http"
	"drassi.run/gha-runner/pkg/lease"
	"drassi.run/gha-runner/pkg/log/logtypes"
	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/timeline"
	util "drassi.run/gha-runner/pkg/types"
)

const (
	taskLogEndpoint        = "/%s/_apis/distributedtask/hubs/%s/plans/%s/logs"
	taskAttachmentEndpoint = "/%s/_apis/distributedtask/hubs/%s/plans/%s/timelines/%s/records/%s/attachments"
	taskLiveFeedEndpoint   = "/%s/_apis/distributedtask/hubs/%s/plans/%s/timelines/%s/records/%s/feed"
	taskTimelineEndpoint   = "/%s/_apis/distributedtask/hubs/%s/plans/%s/timelines/%s/records"
)

type JobService interface {
	LogsUploader(uid string) logtypes.Uploader
	AttachmentUploader(uid, kind, name string) logtypes.Uploader
	LiveFeedAppender() logtypes.Appender
	TimelineRecorder() timeline.Recorder
	WrapLease(l lease.Lease) lease.Lease
}

func NewJobService(url string, hc *http.Client, msg *messages.PipelineAgentJobRequest) (JobService, error) {
	client, err := newClient(url, hc)
	if err != nil {
		return nil, err
	}

	svc := &jobService{
		client:      client,
		scopeUid:    msg.Plan.ScopeIdentifier,
		planType:    msg.Plan.PlanType,
		planUid:     msg.Plan.PlanId,
		timelineUid: msg.Timeline.Id,
	}
	return svc, nil
}

// https://github.com/actions/runner/blob/v2.323.0/src/Runner.Common/JobServer.cs
type jobService struct {
	client      *xhttp.Client
	scopeUid    string // from jobRequest.plan.scopeIdentifier
	planType    string // from jobRequest.plan.planType (a.k.a hubName in C#)
	planUid     string // from jobRequest.plan.planId
	timelineUid string // from jobRequest.timeline.id
}

////////////// Logs //////////////

func (s *jobService) LogsUploader(uid string) logtypes.Uploader {
	return &jobLogsUploader{svc: s, uid: uid}
}

type jobLogsUploader struct {
	svc *jobService
	uid string
}

func (u *jobLogsUploader) Upload(ctx context.Context, r io.Reader, _ *logtypes.Stat) error {
	return u.svc.uploadLogs(ctx, u.uid, r)
}

// https://github.com/actions/runner/blob/v2.323.0/src/Runner.Common/JobServerQueue.cs#L882-L896
func (s *jobService) uploadLogs(ctx context.Context, uid string, r io.Reader) error {
	// Create the log
	tl := new(taskLog)
	endpoint := fmt.Sprintf(taskLogEndpoint, s.scopeUid, s.planType, s.planUid)
	e := s.client.Post(endpoint).
		SetQuery("api-version", "5.1-preview").
		WithBodyProvider(xhttp.JsonEncode(&taskLog{Path: `log\` + uid})).
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

func (s *jobService) AttachmentUploader(uid, kind, name string) logtypes.Uploader {
	return &jobAttachmentUploader{svc: s, uid: uid, kind: kind, name: name}
}

type jobAttachmentUploader struct {
	svc             *jobService
	uid, kind, name string
}

func (u *jobAttachmentUploader) Upload(ctx context.Context, r io.Reader, _ *logtypes.Stat) error {
	return u.svc.uploadAttach(ctx, u.uid, u.kind, u.name, r)
}

// https://github.com/actions/runner/blob/v2.323.0/src/Runner.Common/JobServerQueue.cs#L900-L903
func (s *jobService) uploadAttach(ctx context.Context, uid, kind, name string, r io.Reader) error {
	endpoint := fmt.Sprintf(taskAttachmentEndpoint, s.scopeUid, s.planType, s.planUid, s.timelineUid, uid)
	endpoint = path.Join(endpoint, kind, name)

	e := s.client.Post(endpoint).
		SetQuery("api-version", "5.1-preview").
		SetHeader("Content-Type", "application/octet-stream").
		WithBody(r)

	return e.Do(ctx)
}

////////////// Live Feed //////////////

func (s *jobService) LiveFeedAppender() logtypes.Appender {
	return logtypes.FuncAppender(s.feedingLogs)
}

// https://github.com/actions/runner/blob/v2.332.0/src/Runner.Common/JobServer.cs#L285-L295
// https://github.com/actions/runner/blob/v2.332.0/src/Sdk/DTGenerated/Generated/TaskHttpClientBase.cs#L115-L141
func (s *jobService) feedingLogs(ctx context.Context, uid string, startAt int, lines []string) error {
	data := &logtypes.LinesWrapper{
		Value:     lines,
		Count:     len(lines),
		StepId:    uid,
		StartLine: startAt,
	}

	endpoint := fmt.Sprintf(taskLiveFeedEndpoint, s.scopeUid, s.planType, s.planUid, s.timelineUid, uid)
	e := s.client.Post(endpoint).
		SetQuery("api-version", "5.1-preview").
		WithBodyProvider(xhttp.JsonEncode(data))

	return e.Do(ctx)
}

////////////// Timeline Record //////////////

func (s *jobService) TimelineRecorder() timeline.Recorder {
	return &jobTimelineRecorder{svc: s}
}

type jobTimelineRecorder struct {
	svc *jobService
}

func (r *jobTimelineRecorder) Update(ctx context.Context, records ...*timeline.Record) error {
	if len(records) == 0 {
		return nil
	}

	timelineRecords := make([]*record, len(records))
	for i, rec := range records {
		timelineRecords[i] = toTimelineRecord("", r.svc.timelineUid, rec)
	}

	return r.svc.updateTimelineRecord(ctx, timelineRecords)
}

func (s *jobService) updateTimelineRecord(ctx context.Context, records []*record) error {
	endpoint := fmt.Sprintf(taskTimelineEndpoint, s.scopeUid, s.planType, s.planUid, s.timelineUid)
	body := util.NewList(records)

	e := s.client.Patch(endpoint).
		SetQuery("api-version", "5.1-preview").
		WithBodyProvider(xhttp.JsonEncode(body))

	return e.Do(ctx)
}

func toTimelineRecord(parentId, timelineUid string, rec *timeline.Record) *record {
	r := &record{
		Id:         rec.Uid,
		ParentId:   parentId,
		Name:       rec.Name,
		Order:      rec.Order,
		TimelineId: timelineUid,
		StartTime:  rec.StartedAt,
		FinishTime: rec.CompletedAt,
		State:      rec.State,
		Result:     rec.Result,
		Issues:     rec.Issues,
	}

	for _, i := range rec.Issues {
		switch i.Type {
		case cmdtypes.IssueTypeError:
			r.ErrorCount++
		case cmdtypes.IssueTypeWarning:
			r.WarningCount++
		case cmdtypes.IssueTypeNotice:
			r.NoticeCount++
		}
	}

	if rec.State == timeline.StateCompleted {
		r.PercentComplete = 100
	}

	// https://github.com/actions/runner/blob/v2.324.0/src/Runner.Worker/ExecutionContext.cs#L27-L31
	switch rec.Object.(type) {
	case *timeline.JobObject:
		r.Type = "Job"
	case *timeline.StepObject:
		r.Type = "Task"
	}
	return r
}

func (s *jobService) WrapLease(l lease.Lease) lease.Lease {
	return &jobLeaseWrapper{Lease: l, svc: s}
}

type jobLeaseWrapper struct {
	lease.Lease
	svc *jobService
}

func (l *jobLeaseWrapper) Complete(ctx context.Context, record *timeline.Record) error {
	if err := l.svc.completeJob(ctx, record); err != nil {
		return err
	}
	return l.Lease.Complete(ctx, record)
}

func (s *jobService) completeJob(ctx context.Context, record *timeline.Record) error {
	return nil // TODO
}
