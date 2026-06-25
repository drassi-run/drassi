/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"

	"drassi.run/core/pkg/command/cmdtypes"
	exec "drassi.run/core/pkg/executor"
	"drassi.run/gha-runner/pkg/timeline"
)

type IssueReporter struct {
	mgr *timeline.Manager
}

func NewIssueReporter(mgr *timeline.Manager) cmdtypes.Reporter[exec.Milieu] {
	return &IssueReporter{mgr: mgr}
}

func (r *IssueReporter) AddIssue(_ context.Context, res exec.Milieu, iss *cmdtypes.Issue) error {
	spec := res.StepSpec()
	r.mgr.AddIssue(res.Stage(), spec.Uid, iss)
	return nil
}
