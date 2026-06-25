/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire

import (
	"fmt"
	"net/http"

	xdig "drassi.run/core/util/dig"
	"drassi.run/core/wire"
	wire_command "drassi.run/core/wire/command"
	wire_common "drassi.run/core/wire/common"
	wire_runtime "drassi.run/core/wire/runtime"
	wire_scribe "drassi.run/core/wire/scribe"
	wire_stream "drassi.run/core/wire/stream"
	"drassi.run/gha-runner/pkg/messages"
	wire_core "drassi.run/gha-runner/wire/core"
	wire_lease "drassi.run/gha-runner/wire/lease"
	wire_log "drassi.run/gha-runner/wire/log"
	wire_service "drassi.run/gha-runner/wire/service"
	wire_timeline "drassi.run/gha-runner/wire/timeline"
	"go.uber.org/dig"
)

func Synthetic(scope *dig.Scope, msg *messages.PipelineAgentJobRequest, extras ...*wire.Module) error {
	modules := make([]*wire.Module, 0, 4)

	// core modules
	modules = append(modules, wire_common.Module(
		wire_common.UseEmptyFeatureFlags(false),        // use [wire_core.sysVarFlags] instead
		wire_common.ProvideDefaultDossier(false),       // provided in [wire_coer.newDossier]
		wire_common.ProvideDefaultExpressionEnv(false), // provided in [wire_core.expressionEnv]
	))
	modules = append(modules, wire_command.Module(
		wire_command.UseDiscardIssueReporter(false),        // use [command.IssueReporter] instead
		wire_command.UseBlackHoleAttachmentUploader(false), // use [command.xServiceAttacher] instead
	))
	modules = append(modules, wire_runtime.Module())
	modules = append(modules, wire_scribe.Module())
	modules = append(modules, wire_stream.Module())

	// gha modules
	modules = append(modules, wire_core.Module())
	modules = append(modules, wire_log.Module())
	modules = append(modules, wire_timeline.Module())
	modules = append(modules, wire_lease.Module(msg))
	modules = append(modules, wire_service.Module(msg))

	// extras
	modules = append(modules, extras...)

	if err := xdig.Supply(scope, msg); err != nil {
		return fmt.Errorf("provide messages.PipelineAgentJobRequest: %w", err)
	}
	if err := xdig.Supply(scope, http.DefaultClient); err != nil {
		return fmt.Errorf("provide default http.DefaultClient: %w", err)
	}
	return wire.Apply(scope, modules...)
}
