/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_service

import (
	"fmt"
	"net/http"

	"drassi.run/gha-runner/pkg/log/logsubscriber"
	"drassi.run/gha-runner/pkg/messages"
	"drassi.run/gha-runner/pkg/service"
	"go.uber.org/dig"
)

func wireJobService(scope *dig.Scope, msg *messages.PipelineAgentJobRequest, ep *messages.ServiceEndpoint) error {
	fn := func(hc *http.Client) (service.JobService, error) {
		return service.NewJobService(ep.Url, hc, msg)
	}
	if err := scope.Provide(fn); err != nil {
		return fmt.Errorf("provide service.JobService: %w", err)
	}

	if err := scope.Decorate(service.JobService.WrapLease); err != nil {
		return fmt.Errorf("decorate lease.Lease w/ JobService: %w", err)
	}
	if err := scope.Provide(service.JobService.LiveFeedAppender); err != nil {
		return fmt.Errorf("provide 'jobService' logtypes.Appender: %w", err)
	}
	if err := scope.Provide(service.JobService.TimelineRecorder); err != nil {
		return fmt.Errorf("provide 'jobService' timeline.Recorder: %w", err)
	}
	if err := scope.Provide(logsubscriber.NewJobServiceLogsSubscriber, dig.Group(LogSubscribers)); err != nil {
		return fmt.Errorf("provide 'jobService' logtypes.Subscriber: %w", err)
	}
	return nil
}
