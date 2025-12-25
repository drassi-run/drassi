/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package report

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"

	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/report/types"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

func TestJobServiceTestSuite(t *testing.T) {
	suite.Run(t, new(JobServiceTestSuite))
}

type JobServiceTestSuite struct {
	suite.Suite
	ctrl   *gomock.Controller
	mux    *http.ServeMux
	server *httptest.Server
	svc    *jobService

	scopeUid    string
	planType    string
	planUid     string
	timelineUid string
}

func (s *JobServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mux = http.NewServeMux()
	s.server = httptest.NewServer(s.mux)

	s.scopeUid = "scope-id"
	s.planType = "plan-type"
	s.planUid = "plan-id"
	s.timelineUid = "timeline-id"

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
	svc, err := NewJobService(s.server.URL, nil, msg)
	s.Require().NoError(err)
	s.svc = svc.(*jobService)
}

func (s *JobServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
	s.server.Close()
}

func (s *JobServiceTestSuite) TestLogsUploader() {
	uid := "log-uid"
	logId := "123"
	content := "log content"

	logEndpoint := fmt.Sprintf(taskLogEndpoint, s.scopeUid, s.planType, s.planUid)

	// Step 1: Create the log
	s.mux.HandleFunc("POST "+logEndpoint, func(w http.ResponseWriter, r *http.Request) {
		s.Require().Equal("5.1-preview", r.URL.Query().Get("api-version"))

		var req taskLog
		readJsonRequest(s.T(), r, &req)
		s.Require().Equal(`log\`+uid, req.Path)

		resp := taskLog{Id: logId}
		writeJsonResponse(s.T(), w, &resp)
	})

	// Step 2: Upload the contents
	var contentUploaded bool
	s.mux.HandleFunc("POST "+logEndpoint+"/"+logId, func(w http.ResponseWriter, r *http.Request) {
		s.Equal("5.1-preview", r.URL.Query().Get("api-version"))
		s.Equal("application/octet-stream", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		s.Require().NoError(err)
		s.EqualValues(content, body)

		w.WriteHeader(http.StatusOK)
		contentUploaded = true
	})

	u := s.svc.LogsUploader(uid)
	err := u.Upload(s.T().Context(), strings.NewReader(content), nil)
	s.Require().NoError(err)
	s.True(contentUploaded)
}

func (s *JobServiceTestSuite) TestAttachmentUploader() {
	uid := "record-uid"
	kind := "attachment-kind"
	name := "attachment-name"
	content := "attachment content"

	attachEndpoint := fmt.Sprintf(taskAttachmentEndpoint, s.scopeUid, s.planType, s.planUid, s.timelineUid, uid)
	attachEndpoint = path.Join(attachEndpoint, kind, name)

	s.mux.HandleFunc("POST "+attachEndpoint, func(w http.ResponseWriter, r *http.Request) {
		s.Equal("5.1-preview", r.URL.Query().Get("api-version"))
		s.Equal("application/octet-stream", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		s.Require().NoError(err)
		s.EqualValues(content, body)

		w.WriteHeader(http.StatusOK)
	})

	u := s.svc.AttachmentUploader(uid, kind, name)
	err := u.Upload(s.T().Context(), strings.NewReader(content), nil)
	s.Require().NoError(err)
}

func (s *JobServiceTestSuite) TestLiveFeedAppender() {
	uid := "step-uid"
	lines := []string{"line1", "line2"}
	startAt := 10

	feedEndpoint := fmt.Sprintf(taskLiveFeedEndpoint, s.scopeUid, s.planType, s.planUid, s.timelineUid, uid)

	s.mux.HandleFunc("POST "+feedEndpoint, func(w http.ResponseWriter, r *http.Request) {
		s.Equal("5.1-preview", r.URL.Query().Get("api-version"))

		var req types.LinesWrapper
		readJsonRequest(s.T(), r, &req)
		s.Equal(uid, req.StepId)
		s.Equal(lines, req.Value)
		s.Equal(len(lines), req.Count)
		s.Equal(startAt, req.StartLine)

		w.WriteHeader(http.StatusOK)
	})

	a := s.svc.LiveFeedAppender()
	err := a.Append(s.T().Context(), uid, startAt, lines)
	s.Require().NoError(err)

	err = a.Close()
	s.Require().NoError(err)
}
