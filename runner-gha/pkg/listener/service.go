/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package listener

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"

	"drassi.run/core/util/http"
	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/types"
)

type service interface {
	Connect(ctx context.Context, ref *types.RunnerReference) (*Session, func() error, error)
	GetMessage(ctx context.Context, session *Session, os, arch string) (*messages.Message, error)
	DeleteMessage(ctx context.Context, session *Session, messageId int64) error
}

var _ service = (*runnerService)(nil)

type runnerService struct {
	client *xhttp.Client

	lastMessageId int64
}

const (
	groupEndpoint    = "_apis/distributedtask/pools"
	sessionEndpoint  = groupEndpoint + "/%d/sessions"
	messagesEndpoint = groupEndpoint + "/%d/messages"
)

// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/DTGenerated/Generated/TaskAgentHttpClientBase.cs#L744
func (s *runnerService) Connect(ctx context.Context, ref *types.RunnerReference) (*Session, func() error, error) {
	ss := &Session{Runner: ref}
	if hn, err := os.Hostname(); err != nil {
		ss.OwnerName = "RUNNER"
	} else {
		ss.OwnerName = hn
	}

	r := s.client.Post(fmt.Sprintf(sessionEndpoint, ref.GroupId)).
		SetQuery("api-version", "5.1-preview").
		WithBodyProvider(xhttp.JsonEncode(ss)).
		OnSuccess(xhttp.JsonDecode(ss))
	if err := r.Do(ctx); err != nil {
		return nil, nil, err
	}

	cancel := func() error {
		endpoint := path.Join(fmt.Sprintf(sessionEndpoint, ref.GroupId), ss.Id)
		return s.client.Delete(endpoint).
			SetQuery("api-version", "5.1-preview").
			Do(ctx)
	}

	return ss, cancel, nil
}

// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/DTGenerated/Generated/TaskAgentHttpClientBase.cs#L458
func (s *runnerService) GetMessage(ctx context.Context, session *Session, os, arch string) (*messages.Message, error) {
	runner := session.Runner
	r := s.client.Get(fmt.Sprintf(messagesEndpoint, runner.GroupId)).
		SetQuery("api-version", "6.0-preview").
		SetQuery("sessionId", session.Id).
		SetQuery("disableUpdate", strconv.FormatBool(runner.DisableUpdate))
	if os != "" {
		r.SetQuery("os", os)
	}
	if arch != "" {
		r.SetQuery("architecture", arch)
	}
	if status := runner.Status; status != "" {
		r.SetQuery("status", string(status))
	}
	if version := runner.Version; version != "" {
		r.SetQuery("runnerVersion", version)
	}
	if s.lastMessageId > 0 {
		r.SetQuery("lastMessageId", strconv.FormatInt(s.lastMessageId, 10))
	}

	m := new(messages.Message)
	r.AfterResponseReceive(skipEmpty).OnSuccess(xhttp.JsonDecode(m))
	err := r.Do(ctx)

	if err == nil && m.Id > s.lastMessageId {
		s.lastMessageId = m.Id
	}
	return m, err
}

// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/DTGenerated/Generated/TaskAgentHttpClientBase.cs#L420
func (s *runnerService) DeleteMessage(ctx context.Context, session *Session, messageId int64) error {
	runner := session.Runner
	endpoint := path.Join(fmt.Sprintf(messagesEndpoint, runner.GroupId), strconv.FormatInt(messageId, 10))

	return s.client.Delete(endpoint).
		SetQuery("api-version", "6.0-preview").
		SetQuery("sessionId", session.Id).
		Do(ctx)
}

var _ service = (*brokerService)(nil)

type brokerService struct {
	client *xhttp.Client
}

func (s *brokerService) SetUrl(url string) error {
	if client, err := s.client.WithBaseUrl(url); err != nil {
		return err
	} else {
		s.client = client
		return nil
	}
}

// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/WebApi/WebApi/BrokerHttpClient.cs#L144
func (s *brokerService) Connect(ctx context.Context, ref *types.RunnerReference) (*Session, func() error, error) {
	ss := &Session{Runner: ref}
	if hn, err := os.Hostname(); err != nil {
		ss.OwnerName = "RUNNER"
	} else {
		ss.OwnerName = hn
	}

	r := s.client.Post("session").
		WithBodyProvider(xhttp.JsonEncode(ss)).
		OnSuccess(xhttp.JsonDecode(ss))
	if err := r.Do(ctx); err != nil {
		return nil, nil, err
	}

	// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/WebApi/WebApi/BrokerHttpClient.cs#L176
	cancel := func() error {
		return s.client.Delete("session").Do(ctx)
	}

	return ss, cancel, nil
}

// https://github.com/actions/runner/blob/v2.323.0/src/Sdk/WebApi/WebApi/BrokerHttpClient.cs#L59
func (s *brokerService) GetMessage(ctx context.Context, session *Session, os, arch string) (*messages.Message, error) {
	runner := session.Runner
	r := s.client.Get("message").
		SetQuery("sessionId", session.Id).
		SetQuery("disableUpdate", strconv.FormatBool(runner.DisableUpdate))
	if os != "" {
		r.SetQuery("os", os)
	}
	if arch != "" {
		r.SetQuery("architecture", arch)
	}
	if status := runner.Status; status != "" {
		r.SetQuery("status", string(status))
	}
	if version := runner.Version; version != "" {
		r.SetQuery("runnerVersion", version)
	}

	m := new(messages.Message)
	r.AfterResponseReceive(skipEmpty).OnSuccess(xhttp.JsonDecode(m))
	err := r.Do(ctx)

	return m, err
}

func (s *brokerService) DeleteMessage(context.Context, *Session, int64) error {
	return nil // does nothing
}

func skipEmpty(resp *http.Response) (skip bool, err error) {
	return resp.ContentLength == 0, nil
}
