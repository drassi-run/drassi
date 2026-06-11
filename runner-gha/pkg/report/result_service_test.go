/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package report

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"drassi.run/gha-runner/mock/log/logtypes"
	"drassi.run/gha-runner/pkg/log/logtypes"
	"drassi.run/gha-runner/pkg/messages"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type ResultServiceTestSuite struct {
	suite.Suite
	ctrl   *gomock.Controller
	mux    *http.ServeMux
	server *httptest.Server
	svc    *resultService
}

func TestResultServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ResultServiceTestSuite))
}

func (s *ResultServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mux = http.NewServeMux()
	s.server = httptest.NewServer(s.mux)

	msg := &messages.PipelineAgentJobRequest{
		Plan: messages.PlanReference{
			PlanId: "plan-id",
		},
		JobId: "job-id",
	}
	svc, err := NewResultService(s.server.URL, nil, msg)
	s.Require().NoError(err)
	s.svc = svc.(*resultService)
}

func (s *ResultServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
	s.server.Close()
}

func (s *ResultServiceTestSuite) TestStepLogs_GetSignedUrl() {
	stepUid := "step-id"
	s.mux.HandleFunc("POST "+receiverEndpoint+"GetStepLogsSignedBlobURL", func(w http.ResponseWriter, r *http.Request) {
		var req signedUrlStepLogsRequest
		readJsonRequest(s.T(), r, &req)
		s.Equal("plan-id", req.PlanId)
		s.Equal("job-id", req.JobId)
		s.Equal(stepUid, req.StepId)

		resp := signedUrlStepLogsResponse{
			Url:         "http://signed-url",
			StorageType: logtypes.StorageAzureBlob,
		}
		writeJsonResponse(s.T(), w, &resp)
	})

	resp, err := s.svc.getStepLogsSignedUrl(s.T().Context(), stepUid)
	s.Require().NoError(err)
	s.Equal("http://signed-url", resp.GetUrl())
	s.Equal(logtypes.StorageAzureBlob, resp.GetStorageType())
}

func (s *ResultServiceTestSuite) TestStepLogs_CreateMetadata() {
	stepUid := "step-id"
	lineCount := 42
	s.mux.HandleFunc("POST "+receiverEndpoint+"CreateStepLogsMetadata", func(w http.ResponseWriter, r *http.Request) {
		var req metadataStepLogsRequest
		readJsonRequest(s.T(), r, &req)
		s.Equal("plan-id", req.PlanId)
		s.Equal("job-id", req.JobId)
		s.Equal(stepUid, req.StepId)
		s.Equal(lineCount, req.LineCount)

		writeJsonResponse(s.T(), w, &metadataResponse{Ok: true})
	})

	err := s.svc.createStepLogsMetadata(s.T().Context(), stepUid, lineCount)
	s.Require().NoError(err)
}

func (s *ResultServiceTestSuite) TestJobLogs_GetSignedUrl() {
	s.mux.HandleFunc("POST "+receiverEndpoint+"GetJobLogsSignedBlobURL", func(w http.ResponseWriter, r *http.Request) {
		var req signedUrlJobLogsRequest
		readJsonRequest(s.T(), r, &req)
		s.Equal("plan-id", req.PlanId)
		s.Equal("job-id", req.JobId)

		resp := signedUrlJobLogsResponse{
			Url:         "http://job-signed-url",
			StorageType: logtypes.StorageAzureBlob,
		}
		writeJsonResponse(s.T(), w, &resp)
	})

	resp, err := s.svc.getJobLogsSignedUrl(s.T().Context())
	s.Require().NoError(err)
	s.Equal("http://job-signed-url", resp.GetUrl())
	s.Equal(logtypes.StorageAzureBlob, resp.GetStorageType())
}

func (s *ResultServiceTestSuite) TestJobLogs_CreateMetadata() {
	lineCount := 100
	s.mux.HandleFunc("POST "+receiverEndpoint+"CreateJobLogsMetadata", func(w http.ResponseWriter, r *http.Request) {
		var req metadataJobLogsRequest
		readJsonRequest(s.T(), r, &req)
		s.Equal("plan-id", req.PlanId)
		s.Equal("job-id", req.JobId)
		s.Equal(lineCount, req.LineCount)
		s.WithinDuration(time.Now(), req.UploadedAt, 2*time.Second)

		writeJsonResponse(s.T(), w, &metadataResponse{Ok: true})
	})

	err := s.svc.createJobLogsMetadata(s.T().Context(), lineCount)
	s.Require().NoError(err)
}

func (s *ResultServiceTestSuite) TestDiagnosticLogs_GetSignedUrl() {
	s.mux.HandleFunc("POST "+receiverEndpoint+"GetJobDiagLogsSignedBlobURL", func(w http.ResponseWriter, r *http.Request) {
		var req signedUrlDiagnosticLogsRequest
		readJsonRequest(s.T(), r, &req)
		s.Equal("plan-id", req.PlanId)
		s.Equal("job-id", req.JobId)

		resp := signedUrlDiagnosticLogsResponse{
			Url:         "http://diag-signed-url",
			StorageType: logtypes.StorageAzureBlob,
		}
		writeJsonResponse(s.T(), w, &resp)
	})

	resp, err := s.svc.getDiagnosticLogsSignedUrl(s.T().Context())
	s.Require().NoError(err)
	s.Equal("http://diag-signed-url", resp.GetUrl())
	s.Equal(logtypes.StorageAzureBlob, resp.GetStorageType())
}

func (s *ResultServiceTestSuite) TestStepSummary_GetSignedUrl() {
	stepUid := "step-id"
	s.mux.HandleFunc("POST "+receiverEndpoint+"GetStepSummarySignedBlobURL", func(w http.ResponseWriter, r *http.Request) {
		var req signedUrlStepSummaryRequest
		readJsonRequest(s.T(), r, &req)
		s.Equal("plan-id", req.PlanId)
		s.Equal("job-id", req.JobId)
		s.Equal(stepUid, req.StepId)

		resp := signedUrlStepSummaryResponse{
			Url:         "http://summary-signed-url",
			StorageType: logtypes.StorageAzureBlob,
		}
		writeJsonResponse(s.T(), w, &resp)
	})

	resp, err := s.svc.getStepSummarySignedUrl(s.T().Context(), stepUid)
	s.Require().NoError(err)
	s.Equal("http://summary-signed-url", resp.GetUrl())
	s.Equal(logtypes.StorageAzureBlob, resp.GetStorageType())
}

func (s *ResultServiceTestSuite) TestStepSummary_CreateMetadata() {
	stepUid := "step-id"
	size := int64(1024)
	s.mux.HandleFunc("POST "+receiverEndpoint+"CreateStepSummaryMetadata", func(w http.ResponseWriter, r *http.Request) {
		var req metadataStepSummaryRequest
		readJsonRequest(s.T(), r, &req)
		s.Equal("plan-id", req.PlanId)
		s.Equal("job-id", req.JobId)
		s.Equal(stepUid, req.StepId)
		s.Equal(size, req.Size)

		writeJsonResponse(s.T(), w, &metadataResponse{Ok: true})
	})

	err := s.svc.createStepSummaryMetadata(s.T().Context(), stepUid, size)
	s.Require().NoError(err)
}

func (s *ResultServiceTestSuite) TestResultStepLogsConveyor_Run() {
	ctx := s.T().Context()
	mockConv := mock_logtypes.NewMockConveyor(s.ctrl)
	stat := logtypes.NewStat(10, 100)
	mockConv.EXPECT().Run(ctx).Return(stat, nil)

	receivedMetadata := false
	s.mux.HandleFunc("POST "+receiverEndpoint+"CreateStepLogsMetadata", func(w http.ResponseWriter, r *http.Request) {
		receivedMetadata = true
		writeJsonResponse(s.T(), w, &metadataResponse{Ok: true})
	})

	c := &resultStepLogsConveyor{
		Conveyor: mockConv,
		svc:      s.svc,
		stepUid:  "step-id",
	}

	resStat, err := c.Run(ctx)
	s.Require().NoError(err)
	s.Equal(stat, resStat)
	s.True(receivedMetadata)
}

func (s *ResultServiceTestSuite) TestResultJobLogsConveyor_Run() {
	ctx := s.T().Context()
	mockConv := mock_logtypes.NewMockConveyor(s.ctrl)
	stat := logtypes.NewStat(50, 500)

	mockConv.EXPECT().Run(ctx).Return(stat, nil)

	receivedMetadata := false
	s.mux.HandleFunc("POST "+receiverEndpoint+"CreateJobLogsMetadata", func(w http.ResponseWriter, r *http.Request) {
		receivedMetadata = true
		writeJsonResponse(s.T(), w, &metadataResponse{Ok: true})
	})

	c := &resultJobLogsConveyor{
		Conveyor: mockConv,
		svc:      s.svc,
	}

	resStat, err := c.Run(ctx)
	s.Require().NoError(err)
	s.Equal(stat, resStat)
	s.True(receivedMetadata)
}

func (s *ResultServiceTestSuite) TestResultStepSummaryUploader_Upload() {
	ctx := s.T().Context()
	mockUp := mock_logtypes.NewMockUploader(s.ctrl)
	stat := logtypes.NewStat(0, 2048)
	mockUp.EXPECT().Upload(ctx, gomock.Any(), stat).Return(nil)

	receivedMetadata := false
	s.mux.HandleFunc("POST "+receiverEndpoint+"CreateStepSummaryMetadata", func(w http.ResponseWriter, r *http.Request) {
		receivedMetadata = true
		writeJsonResponse(s.T(), w, &metadataResponse{Ok: true})
	})

	u := &resultStepSummaryUploader{
		Uploader: mockUp,
		svc:      s.svc,
		stepUid:  "step-id",
	}

	err := u.Upload(ctx, nil, stat)
	s.Require().NoError(err)
	s.True(receivedMetadata)
}

func writeJsonResponse(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(w).Encode(v))
}

func readJsonRequest(t *testing.T, r *http.Request, v any) {
	t.Helper()
	require.NoError(t, json.NewDecoder(r.Body).Decode(v))
}
