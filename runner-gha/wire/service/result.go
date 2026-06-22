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

func wireResultService(scope *dig.Scope, msg *messages.PipelineAgentJobRequest, ep *messages.ServiceEndpoint) error {
	fn := func(hc *http.Client) (service.ResultService, error) {
		url := ep.Data["ResultsServiceUrl"]
		return service.NewResultService(url, hc, msg)
	}
	if err := scope.Provide(fn); err != nil {
		return fmt.Errorf("provide service.ResultService: %w", err)
	}
	if err := scope.Provide(service.ResultService.TimelineRecorder); err != nil {
		return fmt.Errorf("provide 'resultService' timeline.Recorder: %w", err)
	}
	if err := scope.Provide(logsubscriber.NewResultServiceStepLogsSubscriber, dig.Group(LogSubscribers)); err != nil {
		return fmt.Errorf("provide 'resultService' logtypes.Subscriber: %w", err)
	}
	return nil
}
