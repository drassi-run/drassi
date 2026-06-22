/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_lease

import (
	"fmt"
	"net/http"

	"drassi.run/core/wire"
	"drassi.run/gha-runner/pkg/lease"
	"drassi.run/gha-runner/pkg/messages"
	"go.uber.org/dig"
)

func Module(msg *messages.PipelineAgentJobRequest) *wire.Module {
	fn := func(scope *dig.Scope) error {
		switch typ := msg.MessageType; typ {
		case messages.TypeRunnerJobRequest:
			// https://github.com/actions/runner/blob/v2.335.1/src/Runner.Worker/JobRunner.cs#L67-L70
			return runLease(scope, msg)
		case messages.TypePipelineAgentJobRequest:
			// https://github.com/actions/runner/blob/v2.335.1/src/Runner.Worker/JobRunner.cs#L89-L105
			return runnerLease(scope, msg)
		default:
			return fmt.Errorf("PipelineAgentJobRequest - unknown message type: %s", typ)
		}
	}
	return wire.NewModule("gha/lease", fn)
}

func runLease(scope *dig.Scope, msg *messages.PipelineAgentJobRequest) error {
	fn := func(hc *http.Client) (lease.Lease, error) {
		ep := msg.ServiceEndpoint("SystemVssConnection")
		if ep == nil {
			return nil, fmt.Errorf("SystemVssConnection service endpoint not available")
		}

		runSvc, err := lease.NewRunService(ep.Url, hc)
		if err != nil {
			return nil, err
		}

		l := runSvc.Lease(msg)
		return l, nil
	}
	if err := scope.Provide(fn); err != nil {
		return fmt.Errorf("provide 'RunService' Lease: %w", err)
	}
	return nil
}

func runnerLease(scope *dig.Scope, msg *messages.PipelineAgentJobRequest) error {
	fn := func(svc *lease.RunnerService) lease.Lease {
		return svc.Lease(msg)
	}

	if err := scope.Provide(fn); err != nil {
		return fmt.Errorf("provide 'RunnerService' Lease: %w", err)
	}
	return nil
}
