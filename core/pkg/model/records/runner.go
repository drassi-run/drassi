/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package records

import "drassi.run/core/pkg/model"

// The `runner` context contains information about the runner that is executing the current job.
// https://docs.github.com/en/actions/learn-github-actions/contexts#runner-context
type RunnerInfo struct {
	Name        string             `json:"name"`
	Os          model.Machine      `json:"os"`
	Arch        model.Architecture `json:"arch"`
	Environment string             `json:"environment"`
	Temp        string             `json:"temp"`
	ToolCache   string             `json:"tool_cache"`
	Workspace   string             `json:"workspace"`
	Debug       string             `json:"debug"`
}
