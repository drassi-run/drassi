package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"time"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/util/context"
	"drassi.run/core/util/http"
	"drassi.run/core/util/reactive"
	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/reporter/store"
	"drassi.run/gha-runner/pkg/types"
	"github.com/coder/websocket"
)

var (
	receiverEndpoint = "twirp/results.services.receiver.Receiver/"
	workflowEndpoint = "twirp/github.actions.results.api.v1.WorkflowStepUpdateService/"
)

func NewResultService(url string, hc *http.Client, msg *messages.PipelineAgentJobRequest) (*ResultService, error) {
	client, err := newClient(url, hc)
	if err != nil {
		return nil, err
	}

	svc := &ResultService{
		client:  client,
		planUid: msg.Plan.PlanId,
		jobUid:  msg.JobId,
		//uploader // TODO
	}
	return svc, nil
}

// https://github.com/actions/runner/blob/v2.323.0/src/Runner.Common/ResultsServer.cs#L20
type ResultService struct {
	client  *xhttp.Client
	planUid string // from jobRequest.plan.planId
	jobUid  string // from jobRequest.jobId

	sm store.Manager
}

////////////// Step Logs //////////////

// StepLogsUploader return Uploader used to handle step logs upload
// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/WebApi/WebApi/ResultsHttpClient.cs#L454
func (s *ResultService) StepLogsUploader(sr executor.StepRun) Uploader {
	return &stepLogsResultUploader{svc: s, sr: sr}
}

type stepLogsResultUploader struct {
	svc *ResultService
	sr  executor.StepRun
}

func (u *stepLogsResultUploader) Upload(ctx context.Context, r io.Reader) error {
	url, o, err := u.svc.getStepLogsSignedUrl(ctx, u.sr.StepId())
	if err != nil {
		return err
	}

	s := u.svc.sm.Get(o)
	if s == nil {
		return fmt.Errorf("no store found for %s", o)
	}

	return s.Put(ctx, url, r)
}

func (s *ResultService) getStepLogsSignedUrl(ctx context.Context, stepUid string) (string, string, error) {
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
		return "", "", err
	}
	if resp.Url == "" {
		return "", "", fmt.Errorf("StepLogs upload failed with empty url")
	}
	return resp.Url, resp.StorageType, nil
}

func (u *stepLogsResultUploader) Complete(ctx context.Context, lineCount int64) error {
	return u.svc.createStepLogsMetadata(ctx, u.sr.StepId(), lineCount)
}

func (s *ResultService) createStepLogsMetadata(ctx context.Context, stepUid string, lineCount int64) error {
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

// JobLogsUploader return Uploader used to handle job logs upload
// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/WebApi/WebApi/ResultsHttpClient.cs#L479
func (s *ResultService) JobLogsUploader() Uploader {
	return &jobLogsResultUploader{svc: s}
}

type jobLogsResultUploader struct {
	svc *ResultService
}

func (u *jobLogsResultUploader) Upload(ctx context.Context, r io.Reader) error {
	url, o, err := u.svc.getJobLogsSignedUrl(ctx)
	if err != nil {
		return err
	}

	s := u.svc.sm.Get(o)
	if s == nil {
		return fmt.Errorf("no store found for %s", o)
	}

	return s.Put(ctx, url, r)
}

func (s *ResultService) getJobLogsSignedUrl(ctx context.Context) (string, string, error) {
	req := &signedUrlJobLogsRequest{
		PlanId: s.planUid,
		JobId:  s.jobUid,
	}
	resp := new(signedUrlJobLogsResponse)
	e := s.client.Post(path.Join(receiverEndpoint, "GetJobLogsSignedBlobURL")).
		WithBodyProvider(xhttp.JsonEncode(req)).
		OnSuccess(xhttp.JsonDecode(resp))

	if err := e.Do(ctx); err != nil {
		return "", "", err
	}
	if resp.Url == "" {
		return "", "", fmt.Errorf("JobLogs upload failed with empty url")
	}
	return resp.Url, resp.StorageType, nil
}

func (u *jobLogsResultUploader) Complete(ctx context.Context, lineCount int64) error {
	return u.svc.createJobLogsMetadata(ctx, lineCount)
}

func (s *ResultService) createJobLogsMetadata(ctx context.Context, lineCount int64) error {
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
	return &diagnosticLogsResultUploader{svc: s}
}

type diagnosticLogsResultUploader struct {
	svc *ResultService
}

func (u *diagnosticLogsResultUploader) Upload(ctx context.Context, r io.Reader) error {
	url, o, err := u.svc.getDiagnosticLogsSignedUrl(ctx)
	if err != nil {
		return err
	}

	s := u.svc.sm.Get(o)
	if s == nil {
		return fmt.Errorf("no store found for %s", o)
	}

	return s.Put(ctx, url, r)
}

func (s *ResultService) getDiagnosticLogsSignedUrl(ctx context.Context) (string, string, error) {
	req := &signedUrlDiagnosticLogsRequest{
		PlanId: s.planUid,
		JobId:  s.jobUid,
	}
	resp := new(signedUrlDiagnosticLogsResponse)
	e := s.client.Post(path.Join(receiverEndpoint, "GetJobDiagLogsSignedBlobURL")).
		WithBodyProvider(xhttp.JsonEncode(req)).
		OnSuccess(xhttp.JsonDecode(resp))

	if err := e.Do(ctx); err != nil {
		return "", "", err
	}
	if resp.Url == "" {
		return "", "", fmt.Errorf("DiagnosticLogs upload failed with empty url")
	}
	return resp.Url, resp.StorageType, nil
}

func (u *diagnosticLogsResultUploader) Complete(ctx context.Context, lineCount int64) error {
	return nil
}

////////////// Step Summary //////////////

// StepSummaryUploader return Uploader used to handle step summary upload
// https://github.com/actions/runner/blob/v2.324.0/src/Sdk/WebApi/WebApi/ResultsHttpClient.cs#L398
func (s *ResultService) StepSummaryUploader(stepUid string) Uploader {
	return &stepSummaryResultUploader{svc: s, stepUid: stepUid}
}

type stepSummaryResultUploader struct {
	svc     *ResultService
	stepUid string
}

func (u *stepSummaryResultUploader) Upload(ctx context.Context, r io.Reader) error {
	url, o, err := u.svc.getStepSummarySignedUrl(ctx, u.stepUid)
	if err != nil {
		return err
	}

	s := u.svc.sm.Get(o)
	if s == nil {
		return fmt.Errorf("no store found for %s", o)
	}

	return s.Put(ctx, url, r)
}

func (s *ResultService) getStepSummarySignedUrl(ctx context.Context, stepUid string) (string, string, error) {
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
		return "", "", err
	}
	if resp.Url == "" {
		return "", "", fmt.Errorf("StepSummary upload failed with empty url")
	}
	return resp.Url, resp.StorageType, nil
}

func (u *stepSummaryResultUploader) Complete(ctx context.Context, size int64) error {
	return u.svc.createStepSummaryMetadata(ctx, u.stepUid, size) // TODO size
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

////////////// Live Feed //////////////

func (s *ResultService) LiveFeeder(contextual xcontext.Provider, wsUrl string) (LiveFeeder, error) {
	ctx := contextual.Context()
	opts := &websocket.DialOptions{
		HTTPClient: s.client.HttpClient(),
	}

	conn, _, err := websocket.Dial(ctx, wsUrl, opts)
	if err != nil {
		return nil, err
	}

	lf := &liveFeeder{
		batcher: reactive.NewThrottleBatcher[*line](100, 500*time.Millisecond),
		SendFn: func(data *linesWrapper) error {
			ctx := contextual.Context()

			if payload, err := json.Marshal(data); err != nil {
				return err
			} else {
				return conn.Write(ctx, websocket.MessageText, payload)
			}
		},
		CloseFn: func() error {
			return conn.Close(websocket.StatusNormalClosure, "bye")
		},
	}
	return lf, nil
}

////////////// Timeline Record //////////////

func (s *ResultService) TimelineRecorder() TimelineRecorder {
	return &resultTimelineRecorder{svc: s}
}

type resultTimelineRecorder struct {
	svc   *ResultService
	order int
}

func (r *resultTimelineRecorder) Update(ctx context.Context, records ...*types.Record) error {
	if len(records) == 0 {
		return nil
	}

	steps := make([]*step, len(records))
	for _, rec := range records {
		s := toStep(rec)
		steps = append(steps, s)
	}

	err := r.svc.updateWorkflowSteps(ctx, r.order, steps)
	r.order++

	return err
}

// https://github.com/actions/runner/blob/v2.324.0/src/Sdk/WebApi/WebApi/ResultsHttpClient.cs#L567
func (s *ResultService) updateWorkflowSteps(ctx context.Context, order int, steps []*step) error {
	req := &stepsUpdateRequest{
		PlanId:      s.planUid,
		JobId:       s.jobUid,
		ChangeOrder: order,
		Steps:       steps,
	}
	resp := new(metadataResponse)
	e := s.client.Post(path.Join(workflowEndpoint, "WorkflowStepsUpdate")).
		WithBodyProvider(xhttp.JsonEncode(req)).
		OnSuccess(xhttp.JsonDecode(resp))

	if err := e.Do(ctx); err != nil {
		return err
	}
	if !resp.Ok {
		return fmt.Errorf("failed to update WorkflowSteps")
	}
	return nil
}

// https://github.com/actions/runner/blob/v2.324.0/src/Sdk/WebApi/WebApi/ResultsHttpClient.cs#L515
func toStep(r *types.Record) *step {
	return &step{
		Id:          r.Uid,
		Number:      r.Order,
		Name:        "rec.Name",
		Status:      toStepStatus(r.State),
		Conclusion:  toStepConclusion(r.Result),
		StartedAt:   r.StartedAt,
		CompletedAt: r.CompletedAt,
	}
}

// https://github.com/actions/runner/blob/v2.324.0/src/Sdk/WebApi/WebApi/ResultsHttpClient.cs#L529
func toStepStatus(s types.State) stepStatus {
	switch s {
	case types.StatePending:
		return stepStatusPending
	case types.StateInProgress:
		return stepStatusInProgress
	case types.StateCompleted:
		return stepStatusCompleted
	default:
		return stepStatusUnknown
	}
}

// https://github.com/actions/runner/blob/v2.324.0/src/Sdk/WebApi/WebApi/ResultsHttpClient.cs#L544
func toStepConclusion(r types.Result) stepConclusion {
	switch r {
	case types.ResultSucceeded:
		return stepConclusionSuccess
	case types.ResultFailed:
		return stepConclusionFailure
	case types.ResultCanceled:
		return stepConclusionCancelled
	case types.ResultSkipped:
		return stepConclusionSkipped
	default:
		return stepConclusionUnknown
	}
}
