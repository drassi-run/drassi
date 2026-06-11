/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package timeline

import (
	"context"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/command/cmdtypes"
)

type IssueReporter struct {
	mgr *Manager
}

func NewIssueReporter(mgr *Manager) *IssueReporter {
	return &IssueReporter{mgr: mgr}
}

func (r *IssueReporter) AddIssue(_ context.Context, res executor.Milieu, issue *cmdtypes.Issue) error {
	stepId := res.StepSpec().Uid
	r.mgr.AddIssue(res.Stage(), stepId, issue)
	return nil
}
