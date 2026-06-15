/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package lease

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
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

func TestRunnerServiceSuite(t *testing.T) {
	suite.Run(t, new(RunnerServiceTestSuite))
}

type RunnerServiceTestSuite struct {
	suite.Suite
	mux    *http.ServeMux
	server *httptest.Server
	svc    *RunnerService
	msg    *messages.PipelineAgentJobRequest
}

func (s *RunnerServiceTestSuite) SetupTest() {
	s.mux = http.NewServeMux()
	s.server = httptest.NewServer(s.mux)

	var err error
	s.svc, err = NewRunnerService(s.server.URL, s.server.Client(), 123)
	s.Require().NoError(err)

	s.msg = jobRequest()
}

func (s *RunnerServiceTestSuite) TearDownTest() {
	s.server.Close()
}

// ---- AcquireJob -------------------------------------------------------------

func (s *RunnerServiceTestSuite) TestAcquireJob_Success() {
	t := s.T()
	messageId := "msg-123"

	s.mux.HandleFunc("GET /_apis/distributedtask/runnermessages/{messageId}", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, messageId, r.PathValue("messageId"))
		assert.Equal(t, "6.0-preview", r.URL.Query().Get("api-version"))
		writeJsonResponse(t, w, s.msg)
	})

	got, err := s.svc.AcquireJob(t.Context(), messageId)
	require.NoError(t, err)
	assert.Equal(t, s.msg.JobId, got.JobId)
	assert.Equal(t, s.msg.RequestId, got.RequestId)
}

func (s *RunnerServiceTestSuite) TestAcquireJob_Error() {
	t := s.T()

	s.mux.HandleFunc("GET /_apis/distributedtask/runnermessages/{messageId}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(types.HeaderActivityId, "activity-id-1")
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := s.svc.AcquireJob(t.Context(), "msg-err")
	var actionError *types.ActionsError
	assert.ErrorAs(t, err, &actionError)
	assert.Equal(t, http.StatusInternalServerError, actionError.StatusCode)
	assert.Equal(t, "activity-id-1", actionError.ActivityId)
}

// ---- Lease ------------------------------------------------------------------

func (s *RunnerServiceTestSuite) TestLease_GetMessage() {
	l := s.svc.Lease(s.msg)
	assert.Same(s.T(), s.msg, l.GetMessage())
}

func (s *RunnerServiceTestSuite) TestLease_Renew() {
	t := s.T()

	count := int64(0)
	s.mux.HandleFunc("PATCH /_apis/distributedtask/pools/{groupId}/jobrequests/{requestId}", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, lockToken, r.URL.Query().Get("lockToken"))
		assert.Equal(t, "orch-123", r.Header.Get("X-VSS-OrchestrationId"))

		var req runnerJobRequest
		readJsonRequest(t, r, &req)
		assert.Equal(t, s.msg.RequestId, req.RequestId)

		assert.Equal(t, strconv.Itoa(s.svc.groupId), r.PathValue("groupId"))
		assert.Equal(t, strconv.FormatInt(s.msg.RequestId, 10), r.PathValue("requestId"))

		count++
		ttl := time.Duration(count) * time.Second // renew when 3/4 ttl time pass
		writeJsonResponse(t, w, &runnerJobRequest{
			LockedUntil: new(time.Now().Add(ttl)),
		})
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

func (s *RunnerServiceTestSuite) TestLease_Renew_Error() {
	t := s.T()
	s.mux.HandleFunc("PATCH /_apis/distributedtask/pools/{groupId}/jobrequests/{requestId}", func(w http.ResponseWriter, r *http.Request) {
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

func (s *RunnerServiceTestSuite) TestLease_Complete() {
	t := s.T()

	var gotReq runnerJobRequest
	s.mux.HandleFunc("PATCH /_apis/distributedtask/pools/{groupId}/jobrequests/{requestId}", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, strconv.Itoa(s.svc.groupId), r.PathValue("groupId"))
		assert.Equal(t, strconv.FormatInt(s.msg.RequestId, 10), r.PathValue("requestId"))

		readJsonRequest(t, r, &gotReq)
		w.WriteHeader(http.StatusOK)
	})

	l := s.svc.Lease(s.msg)
	require.NoError(t, l.Complete(t.Context(), mockRecord))

	assert.Equal(t, s.msg.RequestId, gotReq.RequestId)
	assert.Equal(t, timeline.ResultSucceeded, gotReq.Result)
	assert.False(t, gotReq.FinishTime.IsZero())
}

func (s *RunnerServiceTestSuite) TestLease_Complete_Cancel_Renew() {
	t := s.T()
	ctx := t.Context()

	s.mux.HandleFunc("PATCH /_apis/distributedtask/pools/{groupId}/jobrequests/{requestId}", func(w http.ResponseWriter, r *http.Request) {
		writeJsonResponse(t, w, renewJobResponse{LockedUntil: time.Now().Add(time.Hour)})
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
