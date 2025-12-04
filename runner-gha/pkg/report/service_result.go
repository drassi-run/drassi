/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package report

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"time"

	"drassi.run/core/util/http"
	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/types"
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

	client = client.WithDefaultErrorHandler(types.ParseActionsError).
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
func (s *ResultService) StepLogsConveyor(stepUid string) Conveyor {
	c := NewStorageAwareConveyor(func(ctx context.Context) (signedUrlResponse, error) {
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
	Conveyor
	svc     *ResultService
	stepUid string
}

func (c *resultStepLogsConveyor) Run(ctx context.Context) (*Stat, error) {
	if s, err := c.Conveyor.Run(ctx); err != nil {
		return s, err
	} else {
		err = c.svc.createStepLogsMetadata(ctx, c.stepUid, s.Lines)
		return s, err
	}
}

func (s *ResultService) getStepLogsSignedUrl(ctx context.Context, stepUid string) (signedUrlResponse, error) {
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
func (s *ResultService) JobLogsConveyor() Conveyor {
	c := NewStorageAwareConveyor(s.getJobLogsSignedUrl)
	c = &resultJobLogsConveyor{
		Conveyor: c,
		svc:      s,
	}
	return c
}

type resultJobLogsConveyor struct {
	Conveyor
	svc *ResultService
}

func (c *resultJobLogsConveyor) Run(ctx context.Context) (*Stat, error) {
	if s, err := c.Conveyor.Run(ctx); err != nil {
		return s, err
	} else {
		err = c.svc.createJobLogsMetadata(ctx, s.Lines)
		return s, err
	}
}

func (s *ResultService) getJobLogsSignedUrl(ctx context.Context) (signedUrlResponse, error) {
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
func (s *ResultService) DiagnosticLogsUploader() Uploader {
	return NewStorageAwareUploader(s.getDiagnosticLogsSignedUrl)
}

func (s *ResultService) getDiagnosticLogsSignedUrl(ctx context.Context) (signedUrlResponse, error) {
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
