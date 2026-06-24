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
	"drassi.run/gha-runner/pkg/log/logtypes"
	"drassi.run/gha-runner/pkg/service"
	"drassi.run/gha-runner/pkg/types"
)

func NewJobServiceAttacher(svc service.JobService, store types.RecordStore) cmdtypes.Attacher[exec.Milieu] {
	return &jobServiceAttacher{
		svc:   svc,
		store: store,
	}
}

type jobServiceAttacher struct {
	svc   service.JobService
	store types.RecordStore
}

func (a *jobServiceAttacher) Upload(ctx context.Context, res exec.Milieu, att *cmdtypes.Attachment) error {
	if att.Type != cmdtypes.STEP_SUMMARY {
		return nil
	}

	stepUid := res.StepSpec().Uid
	uid := a.store.RecordUid(res.Stage(), stepUid)
	u := a.svc.AttachmentUploader(uid, "Checks.Step.Summary", stepUid)

	return u.Upload(ctx, att.Reader, nil)
}

func NewResultServiceAttacher(svc service.ResultService, store types.RecordStore) cmdtypes.Attacher[exec.Milieu] {
	return &resultServiceAttacher{
		svc:   svc,
		store: store,
	}
}

type resultServiceAttacher struct {
	svc   service.ResultService
	store types.RecordStore
}

func (a *resultServiceAttacher) Upload(ctx context.Context, res exec.Milieu, att *cmdtypes.Attachment) error {
	if att.Type != cmdtypes.STEP_SUMMARY {
		return nil
	}

	stepUid := res.StepSpec().Uid
	uid := a.store.RecordUid(res.Stage(), stepUid)
	u := a.svc.StepSummaryUploader(uid)

	return u.Upload(ctx, att.Reader, new(logtypes.Stat))
}
