/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package listener

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	xhttp "drassi.run/core/util/http"
	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RunnerServiceTestSuite struct {
	suite.Suite
	mux    *http.ServeMux
	server *httptest.Server
	svc    *runnerService
	runner *types.RunnerReference
	ss     *Session
}

func TestRunnerServiceSuite(t *testing.T) {
	suite.Run(t, new(RunnerServiceTestSuite))
}

func (s *RunnerServiceTestSuite) SetupTest() {
	s.mux = http.NewServeMux()
	s.server = httptest.NewServer(s.mux)

	client, err := xhttp.NewClient(s.server.URL)
	s.Require().NoError(err)
	client = client.WithHttpClient(s.server.Client())

	s.svc = &runnerService{
		client: client,
	}

	s.runner = &types.RunnerReference{
		Id:            123,
		Name:          "test-runner",
		GroupId:       123,
		Version:       "v1.2.3",
		Enabled:       true,
		Status:        types.RunnerStatusOnline,
		DisableUpdate: true,
	}

	s.ss = &Session{
		Id:     "session-id",
		Runner: s.runner,
	}
}

func (s *RunnerServiceTestSuite) TearDownTest() {
	s.server.Close()
}

func (s *RunnerServiceTestSuite) TestConnect_Success() {
	t := s.T()
	wantSessionId := "session-uuid"

	s.mux.HandleFunc("POST /_apis/distributedtask/pools/{groupId}/sessions", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, strconv.Itoa(s.runner.GroupId), r.PathValue("groupId"))

		var ss Session
		readJsonRequest(t, r, &ss)
		assert.Equal(t, s.runner, ss.Runner)
		assert.NotEmpty(t, ss.OwnerName)

		ss.Id = wantSessionId
		writeJsonResponse(t, w, ss)
	})

	s.mux.HandleFunc("DELETE /_apis/distributedtask/pools/{groupId}/sessions/{sessionId}", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, strconv.Itoa(s.runner.GroupId), r.PathValue("groupId"))
		assert.Equal(t, wantSessionId, r.PathValue("sessionId"))
		w.WriteHeader(http.StatusOK)
	})

	session, cancel, err := s.svc.Connect(t.Context(), s.runner)
	require.NoError(t, err)
	assert.Equal(t, wantSessionId, session.Id)
	assert.NotNil(t, cancel)
	assert.NoError(t, cancel())
}

func (s *RunnerServiceTestSuite) TestGetMessage_Success() {
	t := s.T()
	wantMessage := &messages.Message{
		Id:   100,
		Type: "RunnerJobRequest",
		Body: "test-body",
	}

	s.mux.HandleFunc("GET /_apis/distributedtask/pools/{groupId}/messages", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, strconv.Itoa(s.runner.GroupId), r.PathValue("groupId"))

		query := r.URL.Query()
		assert.Equal(t, s.ss.Id, query.Get("sessionId"))
		assert.Equal(t, "linux", query.Get("os"))
		assert.Equal(t, "amd64", query.Get("architecture"))
		assert.Equal(t, strconv.FormatBool(s.runner.DisableUpdate), query.Get("disableUpdate"))
		assert.Equal(t, string(s.runner.Status), query.Get("status"))
		assert.Equal(t, s.runner.Version, query.Get("runnerVersion"))

		writeJsonResponse(t, w, wantMessage)
	})

	msg, err := s.svc.GetMessage(t.Context(), s.ss, "linux", "amd64")
	require.NoError(t, err)
	assert.Equal(t, wantMessage.Id, msg.Id)
	assert.Equal(t, wantMessage.Type, msg.Type)
	assert.Equal(t, int64(100), s.svc.lastMessageId)
}

func (s *RunnerServiceTestSuite) TestGetMessage_Empty() {
	t := s.T()
	s.mux.HandleFunc("GET /_apis/distributedtask/pools/{groupId}/messages", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, strconv.Itoa(s.runner.GroupId), r.PathValue("groupId"))
		w.WriteHeader(http.StatusOK)
	})

	msg, err := s.svc.GetMessage(t.Context(), s.ss, "", "")
	require.NoError(t, err)
	assert.True(t, msg.IsEmpty())
}

func (s *RunnerServiceTestSuite) TestDeleteMessage_Success() {
	t := s.T()
	messageId := int64(2025)

	s.mux.HandleFunc("DELETE /_apis/distributedtask/pools/{groupId}/messages/{messageId}", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, strconv.Itoa(s.runner.GroupId), r.PathValue("groupId"))
		assert.Equal(t, strconv.FormatInt(messageId, 10), r.PathValue("messageId"))
		assert.Equal(t, s.ss.Id, r.URL.Query().Get("sessionId"))
		w.WriteHeader(http.StatusOK)
	})

	err := s.svc.DeleteMessage(t.Context(), s.ss, messageId)
	assert.NoError(t, err)
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
