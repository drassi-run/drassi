package report

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
	"drassi.run/gha-runner/pkg/types"
	"github.com/coder/websocket"
)

const (
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
func (s *ResultService) StepLogsConveyor(sr executor.StepRun) Conveyor {
	stepUid := sr.Base().Uid
	c := NewStorageAwareConveyor(func(ctx context.Context) (SignedUrlResponse, error) {
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

func (s *ResultService) getStepLogsSignedUrl(ctx context.Context, stepUid string) (SignedUrlResponse, error) {
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

func (s *ResultService) getJobLogsSignedUrl(ctx context.Context) (SignedUrlResponse, error) {
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
	return NewStorageAwareUploader(s.getDiagnosticLogsSignedUrl)
}

func (s *ResultService) getDiagnosticLogsSignedUrl(ctx context.Context) (SignedUrlResponse, error) {
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
func (s *ResultService) StepSummaryUploader(sr executor.StepRun) Uploader {
	stepUid := sr.Base().Uid
	u := NewStorageAwareUploader(func(ctx context.Context) (SignedUrlResponse, error) {
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
	Uploader
	svc     *ResultService
	stepUid string
}

func (u *resultStepSummaryUploader) Upload(ctx context.Context, r io.Reader, stat *Stat) error {
	if err := u.Uploader.Upload(ctx, r, stat); err != nil {
		return err
	}
	return u.svc.createStepSummaryMetadata(ctx, u.stepUid, stat.Size)
}

func (s *ResultService) getStepSummarySignedUrl(ctx context.Context, stepUid string) (SignedUrlResponse, error) {
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

////////////// Live Feed //////////////

func (s *ResultService) LiveFeeder(contextual xcontext.Provider, wsUrl string) (LiveFeeder, error) {
	opts := &websocket.DialOptions{
		HTTPClient:      s.client.HttpClient(),
		CompressionMode: websocket.CompressionContextTakeover,
	}

	conn, _, err := websocket.Dial(contextual.Context(), wsUrl, opts)
	if err != nil {
		return nil, err
	}

	lf := &liveFeeder{
		batcher: reactive.NewThrottleBatcher[*line](100, 500*time.Millisecond),
		SendFn: func(data *linesWrapper) error {
			if payload, err := json.Marshal(data); err != nil {
				return err
			} else {
				return conn.Write(contextual.Context(), websocket.MessageText, payload)
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
	for i, rec := range records {
		steps[i] = toStep(rec)
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
