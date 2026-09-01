/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire

import (
	"fmt"

	"drassi.run/core/pkg/expression"
	xdig "drassi.run/core/util/dig"
	"drassi.run/core/wire"
	wire_command "drassi.run/core/wire/command"
	wire_common "drassi.run/core/wire/common"
	wire_runtime "drassi.run/core/wire/runtime"
	wire_scribe "drassi.run/core/wire/scribe"
	wire_secret "drassi.run/core/wire/secret"
	wire_stream "drassi.run/core/wire/stream"
	wire_core "drassi.run/gitea-runner/wire/core"
	wire_reporter "drassi.run/gitea-runner/wire/reporter"
	runnerv1 "gitea.dev/actionslib/runner/v1"
	"go.uber.org/dig"
)

func Synthetic(scope *dig.Scope, task *runnerv1.Task, extras ...*wire.Module) error {
	modules := make([]*wire.Module, 0, 4)

	// core modules
	modules = append(modules, wire_common.Module(
		wire_common.AddExpressionOption(
			expression.WithAlias("gitea", "github"), // make `gitea` variable alias to `github`
		),
	))
	modules = append(modules, wire_command.Module())
	modules = append(modules, wire_runtime.Module())
	modules = append(modules, wire_scribe.Module())
	modules = append(modules, wire_secret.Module())
	modules = append(modules, wire_stream.Module(
		wire_stream.ForwardSinkToHandler(true),
	))

	// gitea modules
	modules = append(modules, wire_core.Module())
	modules = append(modules, wire_reporter.Module(task))

	// extras
	modules = append(modules, extras...)

	if err := xdig.Supply(scope, task); err != nil {
		return fmt.Errorf("provide runnerv1.Task: %w", err)
	}
	return wire.Apply(scope, modules...)
}
