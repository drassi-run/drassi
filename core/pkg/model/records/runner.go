/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package records

import "drassi.run/core/pkg/model"

// The `runner` context contains information about the runner that is executing the current job.
// https://docs.github.com/en/actions/learn-github-actions/contexts#runner-context
type Runner struct {
	Name        string             `json:"name" yaml:"name" actions:"name"`
	Os          model.Machine      `json:"os" yaml:"os" actions:"os"`
	Arch        model.Architecture `json:"arch" yaml:"arch" actions:"arch"`
	Environment string             `json:"environment" yaml:"environment" actions:"environment"`
	Temp        string             `json:"temp" yaml:"temp" actions:"temp"`
	ToolCache   string             `json:"tool_cache" yaml:"tool_cache" actions:"tool_cache"`
	Workspace   string             `json:"workspace" yaml:"workspace" actions:"workspace"`
	Debug       string             `json:"debug" yaml:"debug" actions:"debug"`
}
