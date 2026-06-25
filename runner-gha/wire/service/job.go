/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_service

import (
	"fmt"
	"net/http"
	"time"

	"drassi.run/gha-runner/pkg/common"
	"drassi.run/gha-runner/pkg/log/logsubscriber"
	"drassi.run/gha-runner/pkg/log/logtypes"
	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/service"
	"go.uber.org/dig"
)

func wireJobService(scope *dig.Scope, msg *messages.PipelineAgentJobRequest, ep *messages.ServiceEndpoint) error {
	fn := func(hc *http.Client) (service.JobService, error) {
		return service.NewJobService(ep.Url, hc, msg)
	}
	if err := scope.Provide(fn); err != nil {
		return fmt.Errorf("provide JobService: %w", err)
	}

	if err := scope.Decorate(service.JobService.WrapLease); err != nil {
		return fmt.Errorf("decorate lease.Lease w/ JobService: %w", err)
	}
	if err := scope.Provide(service.JobService.LiveFeedAppender, dig.Name("fallback")); err != nil {
		return fmt.Errorf("provide logtypes.Appender from 'JobService': %w", err)
	}
	if err := scope.Decorate(fallbackAppender); err != nil {
		return fmt.Errorf("decorate Appender w/ JobService fallback: %w", err)
	}
	if err := scope.Provide(service.JobService.TimelineRecorder); err != nil {
		return fmt.Errorf("provide timeline.Recorder from 'JobService': %w", err)
	}
	if err := scope.Provide(logsubscriber.NewJobServiceLogsSubscriber, dig.Group(LogSubscribers)); err != nil {
		return fmt.Errorf("provide logtypes.Subscriber from 'JobService': %w", err)
	}
	if err := scope.Provide(common.NewJobServiceAttacher); err != nil {
		return fmt.Errorf("provide cmdtypes.Attacher from 'JobService': %w", err)
	}
	return nil
}

type appenderParams struct {
	dig.In

	Main     logtypes.Appender `optional:"true"` // WebSocket
	Fallback logtypes.Appender `name:"fallback"` // JobService fallback
}

func fallbackAppender(p appenderParams) logtypes.Appender {
	if p.Main == nil {
		return p.Fallback
	}
	return logtypes.NewFallbackAppender(p.Main, p.Fallback, time.Minute)
}
