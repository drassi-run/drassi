package reporter

import (
	"context"
	"fmt"
	"io"
	"time"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/util/http"
)

var (
	receiverEndpoint = "twirp/results.services.receiver.Receiver/"
	workflowEndpoint = "twirp/github.actions.results.api.v1.WorkflowStepUpdateService/"
)

// https://github.com/actions/runner/blob/v2.323.0/src/Runner.Common/ResultsServer.cs#L20
type resultService struct {
	client  *xhttp.Client
	planUid string // from jobRequest.plan.planId
	jobUid  string // from jobRequest.jobId

	uploader Uploader
}

// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/WebApi/WebApi/ResultsHttpClient.cs#L454
func (s *resultService) UploadStepLogs(ctx context.Context, sr executor.StepRun, r io.Reader) error {
	// Get Signed URL
	sResp := new(signedUrlStepLogsResponse)
	e := s.client.Post("GetStepLogsSignedBlobURL").
		WithBodyProvider(xhttp.JsonEncode(&signedUrlStepLogsRequest{
			PlanUid: s.planUid,
			JobUid:  s.jobUid,
			StepUid: sr.Base().Uid,
		})).
		OnSuccess(xhttp.JsonDecode(sResp))
	if err := e.Do(ctx); err != nil {
		return err
	}
	if sResp.Url == "" {
		return fmt.Errorf("StepLogs upload failed with empty url")
	}

	// Uploading file
	if err := s.uploader.Upload(ctx, sResp.Url, r); err != nil {
		return err
	}

	// Send complete message
	mResp := new(metadataResponse)
	e = s.client.Post("CreateStepLogsMetadata").
		WithBodyProvider(xhttp.JsonEncode(&metadataStepLogsRequest{
			PlanUid:    s.planUid,
			JobUid:     s.jobUid,
			StepUid:    sr.Base().Uid,
			UploadedAt: time.Now(),
			LineCount:  0, // TODO
		})).
		OnSuccess(xhttp.JsonDecode(mResp))
	if err := e.Do(ctx); err != nil {
		return err
	}
	if !mResp.Ok {
		return fmt.Errorf("failed to mark StepLogs upload as complete")
	}
	return nil
}

// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/WebApi/WebApi/ResultsHttpClient.cs#L479
func (s *resultService) UploadJobLogs(ctx context.Context, r io.Reader) error {
	// Get Signed URL
	sResp := new(signedUrlJobLogsResponse)
	e := s.client.Post("GetJobLogsSignedBlobURL").
		WithBodyProvider(xhttp.JsonEncode(&signedUrlJobLogsRequest{
			PlanUid: s.planUid,
			JobUid:  s.jobUid,
		})).
		OnSuccess(xhttp.JsonDecode(sResp))
	if err := e.Do(ctx); err != nil {
		return err
	}
	if sResp.Url == "" {
		return fmt.Errorf("JobLogs upload failed with empty url")
	}

	// Uploading file
	if err := s.uploader.Upload(ctx, sResp.Url, r); err != nil {
		return err
	}

	// Send complete message
	mResp := new(metadataResponse)
	e = s.client.Post("CreateJobLogsMetadata").
		WithBodyProvider(xhttp.JsonEncode(&metadataJobLogsRequest{
			PlanUid:    s.planUid,
			JobUid:     s.jobUid,
			UploadedAt: time.Now(),
			LineCount:  0, // TODO
		})).
		OnSuccess(xhttp.JsonDecode(mResp))
	if err := e.Do(ctx); err != nil {
		return err
	}
	if !mResp.Ok {
		return fmt.Errorf("failed to mark JobLogs upload as complete")
	}
	return nil
}

// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/WebApi/WebApi/ResultsHttpClient.cs#L503
func (s *resultService) UploadDiagnosticLogs(ctx context.Context, r io.Reader) error {
	// Get Signed URL
	sResp := new(signedUrlDiagnosticLogsResponse)
	e := s.client.Post("GetJobDiagLogsSignedBlobURL").
		WithBodyProvider(xhttp.JsonEncode(&signedUrlDiagnosticLogsRequest{
			PlanUid: s.planUid,
			JobUid:  s.jobUid,
		})).
		OnSuccess(xhttp.JsonDecode(sResp))
	if err := e.Do(ctx); err != nil {
		return err
	}
	if sResp.Url == "" {
		return fmt.Errorf("DiagnosticLogs upload failed with empty url")
	}

	// Uploading file
	return s.uploader.Upload(ctx, sResp.Url, r)
}

func (s *resultService) RecordTimeline(ctx context.Context, event any) error {
	return nil
}
