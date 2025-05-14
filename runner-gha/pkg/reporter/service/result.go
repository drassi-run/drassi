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
	"github.com/chainguard-dev/clog"
	"github.com/coder/websocket"
)

var (
	receiverEndpoint = "twirp/results.services.receiver.Receiver/"
	workflowEndpoint = "twirp/github.actions.results.api.v1.WorkflowStepUpdateService/"
)

func NewResultService(url string, hc *http.Client, msg *messages.PipelineAgentJobRequest) (*ResultService, error) {
	url = path.Join(url, receiverEndpoint)
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

// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/WebApi/WebApi/ResultsHttpClient.cs#L454
func (s *ResultService) StepLogsUploader(sr executor.StepRun) Uploader {
	return &stepLogsResultUploader{svc: s, sr: sr}
}

func (s *ResultService) getStepLogsSignedUrl(ctx context.Context, stepUid string) (string, string, error) {
	req := &signedUrlStepLogsRequest{
		PlanUid: s.planUid,
		JobUid:  s.jobUid,
		StepUid: stepUid,
	}
	resp := new(signedUrlStepLogsResponse)
	e := s.client.Post("GetStepLogsSignedBlobURL").
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

func (s *ResultService) createStepLogsMetadata(ctx context.Context, stepUid string, lineCount int64) error {
	req := &metadataStepLogsRequest{
		PlanUid:    s.planUid,
		JobUid:     s.jobUid,
		StepUid:    stepUid,
		UploadedAt: time.Now(),
		LineCount:  lineCount,
	}
	resp := new(metadataResponse)
	e := s.client.Post("CreateStepLogsMetadata").
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

// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/WebApi/WebApi/ResultsHttpClient.cs#L479
func (s *ResultService) JobLogsUploader() Uploader {
	return &jobLogsResultUploader{svc: s}
}

func (s *ResultService) getJobLogsSignedUrl(ctx context.Context) (string, string, error) {
	req := &signedUrlJobLogsRequest{
		PlanUid: s.planUid,
		JobUid:  s.jobUid,
	}
	resp := new(signedUrlJobLogsResponse)
	e := s.client.Post("GetJobLogsSignedBlobURL").
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

func (s *ResultService) createJobLogsMetadata(ctx context.Context, lineCount int64) error {
	req := &metadataJobLogsRequest{
		PlanUid:    s.planUid,
		JobUid:     s.jobUid,
		UploadedAt: time.Now(),
		LineCount:  lineCount,
	}
	resp := new(metadataResponse)
	e := s.client.Post("CreateJobLogsMetadata").
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

// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/WebApi/WebApi/ResultsHttpClient.cs#L503
func (s *ResultService) DiagnosticLogsUploader() Uploader {
	return &diagnosticLogsResultUploader{svc: s}
}

func (s *ResultService) getDiagnosticLogsSignedUrl(ctx context.Context) (string, string, error) {
	req := &signedUrlDiagnosticLogsRequest{
		PlanUid: s.planUid,
		JobUid:  s.jobUid,
	}
	resp := new(signedUrlDiagnosticLogsResponse)
	e := s.client.Post("GetJobDiagLogsSignedBlobURL").
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

func (s *ResultService) StepSummaryUploader(stepUid string) Uploader {
	return &stepSummaryResultUploader{svc: s, stepUid: stepUid}
}

func (s *ResultService) LiveFeeder(contextual xcontext.Provider, wsUrl string) (LiveFeeder, error) {
	ctx := contextual.Context()
	opts := &websocket.DialOptions{
		HTTPClient: s.client.HttpClient(),
	}

	conn, _, err := websocket.Dial(ctx, wsUrl, opts)
	if err != nil {
		return nil, err
	}

	batcher := reactive.NewThrottleBatcher[*line](100, 500*time.Millisecond)

	lf := &resultLiveFeeder{
		conn:       conn,
		batcher:    batcher,
		contextual: contextual,
	}
	return lf, nil
}

func (s *ResultService) RecordTimeline(ctx context.Context, event any) error {
	return nil
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

func (u *stepLogsResultUploader) Complete(ctx context.Context, lineCount int64) error {
	return u.svc.createStepLogsMetadata(ctx, u.sr.StepId(), lineCount)
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

func (u *jobLogsResultUploader) Complete(ctx context.Context, lineCount int64) error {
	return u.svc.createJobLogsMetadata(ctx, lineCount)
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

func (u *diagnosticLogsResultUploader) Complete(ctx context.Context, lineCount int64) error {
	return nil
}

type stepSummaryResultUploader struct {
	svc     *ResultService
	stepUid string
}

func (u *stepSummaryResultUploader) Upload(ctx context.Context, r io.Reader) error {
	//TODO implement me
	panic("implement me")
}

func (u *stepSummaryResultUploader) Complete(ctx context.Context, lineCount int64) error {
	//TODO implement me
	panic("implement me")
}

type resultLiveFeeder struct {
	conn       *websocket.Conn
	batcher    reactive.Batcher[*line]
	contextual xcontext.Provider
	logOffset  int64
}

// https://github.com/actions/runner/blob/v2.324.0/src/Runner.Common/ResultsServer.cs#L220
func (lf *resultLiveFeeder) Handle(_ context.Context, s string) error {
	l := &line{
		stepUid: "TODO", // TODO add stepUid
		number:  lf.logOffset,
		content: s,
	}
	lf.logOffset++

	return lf.batcher.Put(l)
}

func (lf *resultLiveFeeder) Start() error {
	return lf.batcher.Start(lf.send)
}

func (lf *resultLiveFeeder) Close() error {
	lf.batcher.Stop()
	return lf.conn.Close(websocket.StatusNormalClosure, "bye")
}

// https://github.com/actions/runner/blob/v2.324.0/src/Runner.Common/ResultsServer.cs#L220
func (lf *resultLiveFeeder) send(lines []*line) {
	ctx := lf.contextual.Context()

	var (
		stepUid string
		offset  int64
		msg     []string
	)

	// split lines into segments by stepUid
	var prev *line
	for _, curr := range lines {
		if prev != nil && prev.stepUid == curr.stepUid {
			msg = append(msg, curr.content)
			prev = curr
			continue
		}

		// curr is start of a new segment
		// => process the previous segment
		if err := lf.sendE(ctx, stepUid, msg, offset); err != nil {
			clog.Errorf("failed to upload logs: %v", err)
		}

		// save state of a new segment
		stepUid, offset = curr.stepUid, curr.number
		msg = []string{curr.content}
		prev = curr
	}

	// process the last segment
	if err := lf.sendE(ctx, stepUid, msg, offset); err != nil {
		clog.Errorf("failed to upload logs: %v", err)
	}

	return
}

func (lf *resultLiveFeeder) sendE(ctx context.Context, stepUid string, lines []string, offset int64) error {
	if len(lines) == 0 {
		return nil
	}

	data := &liveFeed{
		StepUid:   stepUid,
		Lines:     lines,
		Count:     len(lines),
		StartLine: offset,
	}
	if w, err := lf.conn.Writer(ctx, websocket.MessageText); err != nil {
		return err
	} else if err = json.NewEncoder(w).Encode(data); err != nil {
		_ = w.Close()
		return err
	} else {
		return w.Close()
	}
}
