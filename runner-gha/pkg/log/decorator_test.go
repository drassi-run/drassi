/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package log

import (
	"context"
	"testing"

	mock_executor "drassi.run/core/mock/executor"
	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/model/workflows"
	mock_types "drassi.run/gha-runner/mock/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestDecorator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tempDir := t.TempDir()

	mgr, err := NewManager(tempDir, 1024)
	require.NoError(t, err)
	defer mgr.Close()

	mockStore := mock_types.NewMockRecordStore(ctrl)
	dec := NewDecorator(mgr, mockStore)

	t.Run("job step", func(t *testing.T) {
		executed := false
		var capturedUid string

		mockJobExec := mock_executor.NewMockJobExecutor(ctrl)
		mockJobExec.EXPECT().JobSpec().Return(&executor.JobSpec{Id: "job-1", Uid: "job-uid-1"}).AnyTimes()

		mockStepExec := mock_executor.NewMockStepExecutor(ctrl)
		mockStepExec.EXPECT().JobExecutor().Return(mockJobExec).AnyTimes()

		mockStore.EXPECT().RecordUid(executor.StageMain, "job-uid-1").Return("main_job-uid-1")

		jobTask := &executor.StepTask{
			Stage:    executor.StageMain,
			Kind:     workflows.StepKindJob,
			Executor: mockStepExec,
			Run: func(ctx context.Context) (*records.StepResult, error) {
				executed = true
				var ok bool
				capturedUid, ok = StepUidFromContext(ctx)
				assert.True(t, ok)
				assert.Equal(t, "main_job-uid-1", capturedUid)
				// verify writing to manager with captured UID succeeds
				err := mgr.Handle(capturedUid, "hello from job step")
				assert.NoError(t, err)
				return &records.StepResult{Outcome: records.ResultSuccess}, nil
			},
		}

		run := dec.DecorateStepRun(jobTask)
		res, err := run(context.Background())
		require.NoError(t, err)
		assert.True(t, executed)
		assert.NotNil(t, res)
		assert.Equal(t, records.ResultSuccess, res.Outcome)

		// After step finishes, session should be closed and removed from active
		mgr.mu.RLock()
		assert.NotContains(t, mgr.sessions, capturedUid)
		mgr.mu.RUnlock()
	})

	t.Run("action step", func(t *testing.T) {
		executed := false
		var capturedUid string

		mockStepExec := mock_executor.NewMockStepExecutor(ctrl)
		mockStepExec.EXPECT().StepSpec().Return(&executor.StepSpec{Id: "step-1", Uid: "step-uid-1"}).AnyTimes()

		mockStore.EXPECT().RecordUid(executor.StagePost, "step-uid-1").Return("post_step-uid-1")

		stepTask := &executor.StepTask{
			Stage:    executor.StagePost,
			Kind:     workflows.StepKindAction,
			Executor: mockStepExec,
			Run: func(ctx context.Context) (*records.StepResult, error) {
				executed = true
				var ok bool
				capturedUid, ok = StepUidFromContext(ctx)
				assert.True(t, ok)
				assert.Equal(t, "post_step-uid-1", capturedUid)
				// verify writing to manager with captured UID succeeds
				err := mgr.Handle(capturedUid, "hello from action step")
				assert.NoError(t, err)
				return &records.StepResult{Outcome: records.ResultSuccess}, nil
			},
		}

		run := dec.DecorateStepRun(stepTask)
		res, err := run(context.Background())
		require.NoError(t, err)
		assert.True(t, executed)
		assert.NotNil(t, res)
		assert.Equal(t, records.ResultSuccess, res.Outcome)

		// After step finishes, session should be closed and removed from active
		mgr.mu.RLock()
		assert.NotContains(t, mgr.sessions, capturedUid)
		mgr.mu.RUnlock()
	})
}
