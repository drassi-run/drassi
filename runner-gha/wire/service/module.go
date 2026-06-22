/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_service

import (
	"fmt"
	"net/http"

	"drassi.run/core/wire"
	"drassi.run/gha-runner/pkg/log/logsubscriber"
	"drassi.run/gha-runner/pkg/log/logtypes"
	"drassi.run/gha-runner/pkg/messages"
	"go.uber.org/dig"
)

const LogSubscribers = "log-subscribers"

func Module(msg *messages.PipelineAgentJobRequest) *wire.Module {
	fn := func(scope *dig.Scope) error {
		ep := msg.ServiceEndpoint("SystemVssConnection")
		if ep == nil {
			return fmt.Errorf("SystemVssConnection service endpoint not available")
		}
		if err := scope.Decorate(ep.OAuth2Client); err != nil {
			return fmt.Errorf("decorate http.Client with OAuth2: %w", err)
		}

		// https://github.com/actions/runner/blob/v2.335.1/src/Runner.Worker/JobRunner.cs#L65
		resultOnly := msg.MessageType == messages.TypeRunnerJobRequest

		// https://github.com/actions/runner/blob/v2.335.1/src/Runner.Common/JobServerQueue.cs#L115-L118
		if !resultOnly {
			if err := wireJobService(scope, msg, ep); err != nil {
				return err
			}
		}

		// https://github.com/actions/runner/blob/v2.335.1/src/Runner.Common/JobServerQueue.cs#L121-L140
		url := ep.Data["ResultsServiceUrl"]
		if url != "" {
			if err := wireResultService(scope, msg, ep); err != nil {
				return err
			}
		}

		wsUrl := ep.Data["FeedStreamUrl"]
		if wsUrl != "" {
			if err := scope.Provide(wsAppender(wsUrl)); err != nil {
				return fmt.Errorf("provide 'websocket' logtypes.Appender: %w", err)
			}
		}

		if err := scope.Provide(logsubscriber.NewLiveFeedSubscriber, dig.Group(LogSubscribers)); err != nil {
			return fmt.Errorf("provide 'live-feed' logtypes.Subscriber: %w", err)
		}

		return nil
	}
	return wire.NewModule("gha/service", fn)
}

func wsAppender(wsUrl string) func(hc *http.Client) logtypes.Appender {
	return func(hc *http.Client) logtypes.Appender {
		return logtypes.NewWebsocketAppender(wsUrl, hc)
	}
}
