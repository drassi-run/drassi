/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package report

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"

	"drassi.run/core/util/http"
	"drassi.run/gha-runner/pkg/messages"
)

const (
	taskLogEndpoint        = "%s/_apis/distributedtask/hubs/%s/plans/%s/logs"
	taskAttachmentEndpoint = "%s/_apis/distributedtask/hubs/%s/plans/%s/timelines/%s/records/%s/attachments"
	taskLiveFeedEndpoint   = "%s/_apis/distributedtask/hubs/%s/plans/%s/timelines/%s/records/%s/feed"
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

func (s *JobService) LogsUploader(uid string) Uploader {
	return &jobLogsUploader{svc: s, uid: uid}
}

type jobLogsUploader struct {
	svc *JobService
	uid string
}

func (u *jobLogsUploader) Upload(ctx context.Context, r io.Reader, _ *Stat) error {
	return u.svc.uploadLogs(ctx, u.uid, r)
}

// https://github.com/actions/runner/blob/v2.323.0/src/Runner.Common/JobServerQueue.cs#L882-L896
func (s *JobService) uploadLogs(ctx context.Context, uid string, r io.Reader) error {
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

func (s *JobService) AttachmentUploader(uid, kind, name string) Uploader {
	return &jobAttachmentUploader{svc: s, uid: uid, kind: kind, name: name}
}

type jobAttachmentUploader struct {
	svc             *JobService
	uid, kind, name string
}

func (u *jobAttachmentUploader) Upload(ctx context.Context, r io.Reader, _ *Stat) error {
	return u.svc.uploadAttach(ctx, u.uid, u.kind, u.name, r)
}

// https://github.com/actions/runner/blob/v2.323.0/src/Runner.Common/JobServerQueue.cs#L900-L903
func (s *JobService) uploadAttach(ctx context.Context, uid, kind, name string, r io.Reader) error {
	endpoint := fmt.Sprintf(taskAttachmentEndpoint, s.scopeUid, s.planType, s.planUid, s.timelineUid, uid)
	endpoint = path.Join(endpoint, kind, name)

	e := s.client.Post(endpoint).
		SetQuery("api-version", "5.1-preview").
		SetHeader("Content-Type", "application/octet-stream").
		WithBody(r)

	return e.Do(ctx)
}

////////////// Live Feed //////////////

func (s *JobService) LiveFeedAppender() Appender {
	return funcAppender(s.feedingLogs)
}

// https://github.com/actions/runner/blob/v2.332.0/src/Runner.Common/JobServer.cs#L285-L295
// https://github.com/actions/runner/blob/v2.332.0/src/Sdk/DTGenerated/Generated/TaskHttpClientBase.cs#L115-L141
func (s *JobService) feedingLogs(ctx context.Context, uid string, startAt int, lines []string) error {
	data := &linesWrapper{
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
