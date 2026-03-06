/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package lease

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/timeline"
	"drassi.run/gha-runner/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ============================================================================
// RunServiceTestSuite — tests for RunService (AcquireJob, Lease)
// ============================================================================

func TestRunServiceSuite(t *testing.T) {
	suite.Run(t, new(RunServiceTestSuite))
}

type RunServiceTestSuite struct {
	suite.Suite
	mux    *http.ServeMux
	server *httptest.Server
	svc    *RunService
	msg    *messages.PipelineAgentJobRequest
}

func (s *RunServiceTestSuite) SetupTest() {
	s.mux = http.NewServeMux()
	s.server = httptest.NewServer(s.mux)

	var err error
	s.svc, err = NewRunService(s.server.URL, s.server.Client())
	s.Require().NoError(err)

	s.msg = jobRequest()
}

func (s *RunServiceTestSuite) TearDownTest() {
	s.server.Close()
}

// ---- AcquireJob -------------------------------------------------------------

func (s *RunServiceTestSuite) TestAcquireJob_Success() {
	t := s.T()
	want := &messages.PipelineAgentJobRequest{
		JobId:          "job-abc",
		BillingOwnerId: "billing-xyz",
	}

	var gotReq acquireJobRequest
	s.mux.HandleFunc("POST /acquirejob", func(w http.ResponseWriter, r *http.Request) {
		readJsonRequest(t, r, &gotReq)
		writeJsonResponse(t, w, want)
	})

	got, err := s.svc.AcquireJob(t.Context(), "msg-id-1", "billing-xyz")
	require.NoError(t, err)

	// Verify request payload sent to server
	assert.Equal(t, "msg-id-1", gotReq.JobMessageId)
	assert.Equal(t, "billing-xyz", gotReq.BillingOwnerId)
	assert.Equal(t, "Linux", gotReq.RunnerOS)

	// Verify response mapped correctly
	assert.Equal(t, want.JobId, got.JobId)
	assert.Equal(t, want.BillingOwnerId, got.BillingOwnerId)
}

func (s *RunServiceTestSuite) TestAcquireJob_ServerError() {
	t := s.T()
	s.mux.HandleFunc("POST /acquirejob", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := s.svc.AcquireJob(t.Context(), "msg-id-err", "billing")
	var actionError *types.ActionsError
	assert.ErrorAs(t, err, &actionError)
	assert.Equal(t, http.StatusInternalServerError, actionError.StatusCode)
}

// ---- Lease / GetMessage -----------------------------------------------------

func (s *RunServiceTestSuite) TestLease_GetMessage() {
	l := s.svc.Lease(s.msg)
	assert.Same(s.T(), s.msg, l.GetMessage())
}

// ---- Lease / Renew ----------------------------------------------------------

func (s *RunServiceTestSuite) TestLease_Renew() {
	t := s.T()

	count := int64(0)
	s.mux.HandleFunc("POST /renewjob", func(w http.ResponseWriter, r *http.Request) {
		var req renewJobRequest
		readJsonRequest(t, r, &req)

		assert.Equal(t, s.msg.Plan.PlanId, req.PlanId)
		assert.Equal(t, s.msg.JobId, req.JobId)

		count++
		ttl := time.Duration(count) * time.Second // renew when 3/4 ttl time pass
		writeJsonResponse(t, w, renewJobResponse{LockedUntil: time.Now().Add(ttl)})
	})

	l := s.svc.Lease(s.msg)
	ctx, cancel := context.WithCancel(t.Context())
	var done atomic.Bool
	go func() {
		l.Renew(ctx)
		done.Store(true)
	}()

	// renew at 0, 1*3/4=0.75s, (1+2)*3/4=2.25s
	// next renew at (1+2+3)*3/4=4.5s <= will not run
	time.Sleep(3 * time.Second)
	cancel()

	time.Sleep(2 * time.Second) // pass 4th renew
	assert.EqualValues(t, 3, count)
	assert.EqualValues(t, true, done.Load())
}

func (s *RunServiceTestSuite) TestLease_Renew_Error() {
	t := s.T()
	s.mux.HandleFunc("POST /renewjob", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	l := s.svc.Lease(s.msg)
	var done atomic.Bool
	go func() {
		l.Renew(t.Context())
		done.Store(true)
	}()

	// goroutine should stop when renew error
	assert.Eventually(t, done.Load, time.Second, 10*time.Millisecond)
}

// ---- Lease / Complete ----------------------------------------------------------

func (s *RunServiceTestSuite) TestLease_Complete() {
	t := s.T()

	var gotReq completeJobRequest
	s.mux.HandleFunc("POST /completejob", func(w http.ResponseWriter, r *http.Request) {
		readJsonRequest(t, r, &gotReq)
		w.WriteHeader(http.StatusOK)
	})

	l := s.svc.Lease(s.msg)
	require.NoError(t, l.Complete(t.Context(), mockRecord))

	assert.Equal(t, s.msg.JobId, gotReq.JobId)
	assert.Equal(t, s.msg.Plan.PlanId, gotReq.PlanId)
	assert.Equal(t, timeline.ResultSucceeded, gotReq.Conclusion)
}

func (s *RunServiceTestSuite) TestLease_Complete_Cancel_Renew() {
	t := s.T()
	ctx := t.Context()

	s.mux.HandleFunc("POST /renewjob", func(w http.ResponseWriter, r *http.Request) {
		writeJsonResponse(t, w, renewJobResponse{LockedUntil: time.Now().Add(time.Hour)})
	})
	s.mux.HandleFunc("POST /completejob", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	l := s.svc.Lease(s.msg)
	var done atomic.Bool
	go func() {
		l.Renew(ctx)
		done.Store(true)
	}()

	time.Sleep(time.Second) // waiting for renew goroutine start
	require.NoError(t, l.Complete(ctx, mockRecord))

	time.Sleep(time.Second)
	assert.True(t, done.Load())
}

// ---- test helpers -----------------------------------------------------------

func writeJsonResponse(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(w).Encode(v))
}

func readJsonRequest(t *testing.T, r *http.Request, v any) {
	t.Helper()
	require.NoError(t, json.NewDecoder(r.Body).Decode(v))
}

func jobRequest() *messages.PipelineAgentJobRequest {
	return &messages.PipelineAgentJobRequest{
		RequestId:      12345,
		JobId:          "job-uuid-1234",
		BillingOwnerId: "billing-owner-99",
		Plan: messages.PlanReference{
			PlanId: "plan-uuid-5678",
		},
		Variables: map[string]messages.Variable{
			"system.orchestrationId": {Value: "orch-123"},
		},
	}
}

var mockRecord = &timeline.Record{
	Uid:    "job-uuid-1234", // equals jobRequest.JobId
	Object: new(timeline.JobObject),
	State:  timeline.StateCompleted,
	Result: timeline.ResultSucceeded,
}
