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
	"time"

	"drassi.run/core/util/http"
	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/report/types"
	util "drassi.run/gha-runner/pkg/types"
)

const (
	receiverEndpoint = "twirp/results.services.receiver.Receiver/"
	workflowEndpoint = "twirp/github.actions.results.api.v1.WorkflowStepUpdateService/"
)

func newClient(url string, hc *http.Client) (*xhttp.Client, error) {
	client, err := xhttp.NewClient(url)
	if err != nil {
		return nil, err
	}

	client = client.WithDefaultErrorHandler(util.ParseActionsError).
		WithDefaultHeader("User-Agent", "gha-runner") // TODO

	if hc != nil {
		client = client.WithHttpClient(hc)
	}
	return client, nil
}

func NewResultService(url string, hc *http.Client, msg *messages.PipelineAgentJobRequest) (*ResultService, error) {
	client, err := newClient(url, hc)
	if err != nil {
		return nil, err
	}

	svc := &ResultService{
		client:  client,
		planUid: msg.Plan.PlanId,
		jobUid:  msg.JobId,
	}
	return svc, nil
}

// https://github.com/actions/runner/blob/v2.323.0/src/Runner.Common/ResultsServer.cs#L20
type ResultService struct {
	client  *xhttp.Client
	planUid string // from jobRequest.plan.planId - UUID
	jobUid  string // from jobRequest.jobId - UUID
}

////////////// Step Logs //////////////

// StepLogsConveyor return Conveyor used to handle step logs upload
// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/WebApi/WebApi/ResultsHttpClient.cs#L454
func (s *ResultService) StepLogsConveyor(stepUid string) types.Conveyor {
	c := types.NewStorageAwareConveyor(func(ctx context.Context) (types.SignedUrlResponse, error) {
		return s.getStepLogsSignedUrl(ctx, stepUid)
	})
	c = &resultStepLogsConveyor{
		Conveyor: c,
		svc:      s,
		stepUid:  stepUid,
	}
	return c
}

type resultStepLogsConveyor struct {
	types.Conveyor
	svc     *ResultService
	stepUid string
}

func (c *resultStepLogsConveyor) Run(ctx context.Context) (*types.Stat, error) {
	if s, err := c.Conveyor.Run(ctx); err != nil {
		return s, err
	} else {
		err = c.svc.createStepLogsMetadata(ctx, c.stepUid, s.Lines)
		return s, err
	}
}

func (s *ResultService) getStepLogsSignedUrl(ctx context.Context, stepUid string) (types.SignedUrlResponse, error) {
	req := &signedUrlStepLogsRequest{
		PlanId: s.planUid,
		JobId:  s.jobUid,
		StepId: stepUid,
	}
	resp := new(signedUrlStepLogsResponse)
	e := s.client.Post(path.Join(receiverEndpoint, "GetStepLogsSignedBlobURL")).
		WithBodyProvider(xhttp.JsonEncode(req)).
		OnSuccess(xhttp.JsonDecode(resp))

	if err := e.Do(ctx); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *ResultService) createStepLogsMetadata(ctx context.Context, stepUid string, lineCount int) error {
	req := &metadataStepLogsRequest{
		PlanId:     s.planUid,
		JobId:      s.jobUid,
		StepId:     stepUid,
		UploadedAt: time.Now(),
		LineCount:  lineCount,
	}
	resp := new(metadataResponse)
	e := s.client.Post(path.Join(receiverEndpoint, "CreateStepLogsMetadata")).
		WithBodyProvider(xhttp.JsonEncode(req)).
		OnSuccess(xhttp.JsonDecode(resp))

	if err := e.Do(ctx); err != nil {
		return err
	}
	if !resp.Ok {
		return fmt.Errorf("failed to mark StepLogs upload as complete")
	}
	return nil
}

////////////// Job Logs //////////////

// JobLogsConveyor return Conveyor used to handle job logs upload
// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/WebApi/WebApi/ResultsHttpClient.cs#L479
func (s *ResultService) JobLogsConveyor() types.Conveyor {
	c := types.NewStorageAwareConveyor(s.getJobLogsSignedUrl)
	c = &resultJobLogsConveyor{
		Conveyor: c,
		svc:      s,
	}
	return c
}

type resultJobLogsConveyor struct {
	types.Conveyor
	svc *ResultService
}

func (c *resultJobLogsConveyor) Run(ctx context.Context) (*types.Stat, error) {
	if s, err := c.Conveyor.Run(ctx); err != nil {
		return s, err
	} else {
		err = c.svc.createJobLogsMetadata(ctx, s.Lines)
		return s, err
	}
}

func (s *ResultService) getJobLogsSignedUrl(ctx context.Context) (types.SignedUrlResponse, error) {
	req := &signedUrlJobLogsRequest{
		PlanId: s.planUid,
		JobId:  s.jobUid,
	}
	resp := new(signedUrlJobLogsResponse)
	e := s.client.Post(path.Join(receiverEndpoint, "GetJobLogsSignedBlobURL")).
		WithBodyProvider(xhttp.JsonEncode(req)).
		OnSuccess(xhttp.JsonDecode(resp))

	if err := e.Do(ctx); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *ResultService) createJobLogsMetadata(ctx context.Context, lineCount int) error {
	req := &metadataJobLogsRequest{
		PlanId:     s.planUid,
		JobId:      s.jobUid,
		UploadedAt: time.Now(),
		LineCount:  lineCount,
	}
	resp := new(metadataResponse)
	e := s.client.Post(path.Join(receiverEndpoint, "CreateJobLogsMetadata")).
		WithBodyProvider(xhttp.JsonEncode(req)).
		OnSuccess(xhttp.JsonDecode(resp))

	if err := e.Do(ctx); err != nil {
		return err
	}
	if !resp.Ok {
		return fmt.Errorf("failed to mark JobLogs upload as complete")
	}
	return nil
}

////////////// Diagnostic Logs //////////////

// DiagnosticLogsUploader return Uploader used to handle diagnostic logs upload
// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/WebApi/WebApi/ResultsHttpClient.cs#L503
func (s *ResultService) DiagnosticLogsUploader() types.Uploader {
	return types.NewStorageAwareUploader(s.getDiagnosticLogsSignedUrl)
}

func (s *ResultService) getDiagnosticLogsSignedUrl(ctx context.Context) (types.SignedUrlResponse, error) {
	req := &signedUrlDiagnosticLogsRequest{
		PlanId: s.planUid,
		JobId:  s.jobUid,
	}
	resp := new(signedUrlDiagnosticLogsResponse)
	e := s.client.Post(path.Join(receiverEndpoint, "GetJobDiagLogsSignedBlobURL")).
		WithBodyProvider(xhttp.JsonEncode(req)).
		OnSuccess(xhttp.JsonDecode(resp))

	if err := e.Do(ctx); err != nil {
		return nil, err
	}
	return resp, nil
}

////////////// Step Summary //////////////

// StepSummaryUploader return Uploader used to handle step summary upload
// https://github.com/actions/runner/blob/v2.324.0/src/Sdk/WebApi/WebApi/ResultsHttpClient.cs#L398
func (s *ResultService) StepSummaryUploader(stepUid string) types.Uploader {
	u := types.NewStorageAwareUploader(func(ctx context.Context) (types.SignedUrlResponse, error) {
		return s.getStepSummarySignedUrl(ctx, stepUid)
	})
	u = &resultStepSummaryUploader{
		Uploader: u,
		svc:      s,
		stepUid:  stepUid,
	}
	return u
}

type resultStepSummaryUploader struct {
	types.Uploader
	svc     *ResultService
	stepUid string
}

func (u *resultStepSummaryUploader) Upload(ctx context.Context, r io.Reader, stat *types.Stat) error {
	if err := u.Uploader.Upload(ctx, r, stat); err != nil {
		return err
	}
	return u.svc.createStepSummaryMetadata(ctx, u.stepUid, stat.Size)
}

func (s *ResultService) getStepSummarySignedUrl(ctx context.Context, stepUid string) (types.SignedUrlResponse, error) {
	req := &signedUrlStepSummaryRequest{
		PlanId: s.planUid,
		JobId:  s.jobUid,
		StepId: stepUid,
	}
	resp := new(signedUrlStepSummaryResponse)
	e := s.client.Post(path.Join(receiverEndpoint, "GetStepSummarySignedBlobURL")).
		WithBodyProvider(xhttp.JsonEncode(req)).
		OnSuccess(xhttp.JsonDecode(resp))

	if err := e.Do(ctx); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *ResultService) createStepSummaryMetadata(ctx context.Context, stepUid string, size int64) error {
	req := &metadataStepSummaryRequest{
		PlanId:     s.planUid,
		JobId:      s.jobUid,
		StepId:     stepUid,
		UploadedAt: time.Now(),
		Size:       size,
	}
	resp := new(metadataResponse)
	e := s.client.Post(path.Join(receiverEndpoint, "CreateStepSummaryMetadata")).
		WithBodyProvider(xhttp.JsonEncode(req)).
		OnSuccess(xhttp.JsonDecode(resp))

	if err := e.Do(ctx); err != nil {
		return err
	}
	if !resp.Ok {
		return fmt.Errorf("failed to mark StepSummary upload as complete")
	}
	return nil
}
