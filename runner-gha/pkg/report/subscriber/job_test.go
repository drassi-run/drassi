/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"testing/synctest"

	xcontext "drassi.run/core/util/context"
	mock_report "drassi.run/gha-runner/mock/report"
	mock_types "drassi.run/gha-runner/mock/report/types"
	"drassi.run/gha-runner/pkg/log"
	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/report"
	"drassi.run/gha-runner/pkg/report/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

func TestJobServiceLogsSubscriberSuite(t *testing.T) {
	suite.Run(t, new(JobServiceLogsSubscriberTestSuite))
}

type JobServiceLogsSubscriberTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller
	svc  *mock_report.MockJobService
	sub  *jobServiceLogsSubscriber

	tmpDir string

	scopeUid    string
	planType    string
	planUid     string
	timelineUid string
}

func (s *JobServiceLogsSubscriberTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())

	s.svc = mock_report.NewMockJobService(s.ctrl)
	ctx := xcontext.NewStaticProvider(s.T().Context())
	s.sub = NewJobServiceLogsSubscriber(ctx, s.svc).(*jobServiceLogsSubscriber)
	s.tmpDir = s.T().TempDir()

	s.scopeUid = "scope-id"
	s.planType = "plan-type"
	s.planUid = "plan-id"
	s.timelineUid = "timeline-id"
}

func (s *JobServiceLogsSubscriberTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *JobServiceLogsSubscriberTestSuite) TestRun() {
	synctest.Test(s.T(), func(t *testing.T) {
		content := "log line 1\n"
		logFile := s.tempFile("job.log", content)

		uid := "test-uid"
		event := &log.Event{
			Uid: uid,
			Update: &log.Update{
				File:     logFile,
				Line:     1,
				Offset:   int64(len(content)),
				Complete: true,
			},
		}

		u := types.FuncUploader(func(ctx context.Context, r io.Reader, stat *types.Stat) error {
			s.Equal(1, stat.Lines)
			s.EqualValues(len(content), stat.Size)

			data, err := io.ReadAll(r)
			s.Require().NoError(err)
			s.EqualValues(content, data)
			return nil
		})

		s.svc.EXPECT().LogsUploader(uid).
			Return(u).
			AnyTimes()

		ch := make(chan *log.Event)

		go s.sub.Run(ch)
		ch <- event
		close(ch)

		s.sub.Wait()
	})
}

func (s *JobServiceLogsSubscriberTestSuite) TestUploaderCaching() {
	callCount := 0
	mockUploader := mock_types.NewMockUploader(s.ctrl)
	s.svc.EXPECT().LogsUploader(gomock.Any()).
		DoAndReturn(func(uid string) types.Uploader {
			callCount++
			return mockUploader
		}).AnyTimes()

	u1 := s.sub.uploader("uid1")
	s.Equal(mockUploader, u1)
	s.Equal(1, callCount)

	u2 := s.sub.uploader("uid1")
	s.Equal(mockUploader, u2)
	s.Equal(1, callCount, "Should return cached uploader")

	u3 := s.sub.uploader("uid2")
	s.Equal(mockUploader, u3)
	s.Equal(2, callCount, "Should call provider for new uid")
}

func (s *JobServiceLogsSubscriberTestSuite) TestIntegrationWithService() {
	ctx := s.T().Context()
	cp := xcontext.NewStaticProvider(ctx)

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	msg := &messages.PipelineAgentJobRequest{
		Plan: messages.PlanReference{
			ScopeIdentifier: s.scopeUid,
			PlanType:        s.planType,
			PlanId:          s.planUid,
		},
		Timeline: messages.TimelineReference{
			Id: s.timelineUid,
		},
	}
	svc, err := report.NewJobService(server.URL, nil, msg)
	s.Require().NoError(err)

	sub := NewJobServiceLogsSubscriber(cp, svc)
	ch := make(chan *log.Event)

	uid := "log-uid"
	logId := "123"
	content := "log content"
	logFile := s.tempFile("job_integration.log", content)

	logEndpoint := fmt.Sprintf("/%s/_apis/distributedtask/hubs/%s/plans/%s/logs", s.scopeUid, s.planType, s.planUid)

	// Step 1: Create the log
	mux.HandleFunc("POST "+logEndpoint, func(w http.ResponseWriter, r *http.Request) {
		s.Require().Equal("5.1-preview", r.URL.Query().Get("api-version"))

		var req struct {
			Path string `json:"path"`
		}
		readJsonRequest(s.T(), r, &req)
		s.Require().Equal(`log\`+uid, req.Path)

		resp := struct {
			Id string `json:"id"`
		}{Id: logId}
		writeJsonResponse(s.T(), w, &resp)
	})

	// Step 2: Upload the contents
	var contentUploaded bool
	mux.HandleFunc("POST "+logEndpoint+"/"+logId, func(w http.ResponseWriter, r *http.Request) {
		s.Equal("5.1-preview", r.URL.Query().Get("api-version"))
		s.Equal("application/octet-stream", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		s.Require().NoError(err)
		s.EqualValues(content, body)

		w.WriteHeader(http.StatusOK)
		contentUploaded = true
	})

	event := &log.Event{
		Uid: uid,
		Update: &log.Update{
			File:     logFile,
			Line:     1,
			Offset:   int64(len(content)),
			Complete: true,
		},
	}

	go sub.Run(ch)
	ch <- event
	close(ch)

	sub.Wait()
	s.True(contentUploaded)
}

func (s *JobServiceLogsSubscriberTestSuite) tempFile(name, content string) string {
	f := filepath.Join(s.tmpDir, name)
	err := os.WriteFile(f, []byte(content), 0644)
	s.Require().NoError(err)
	return f
}

// Helpers

func writeJsonResponse(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(w).Encode(v))
}

func readJsonRequest(t *testing.T, r *http.Request, v any) {
	t.Helper()
	require.NoError(t, json.NewDecoder(r.Body).Decode(v))
}
