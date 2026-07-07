/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package redact

import (
	"context"
	"testing"

	mock_secret "drassi.run/core/mock/secret"
	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/model/records"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestJobOutputs(t *testing.T) {
	ctrl := gomock.NewController(t)
	sm := mock_secret.NewMockMasker(ctrl)

	result := &records.JobResult{
		Result: records.ResultFailure,
		Outputs: map[string]string{
			"clean": "public value",
			"token": "secret-token",
		},
	}

	sm.EXPECT().IsClean("public value").Return(true)
	sm.EXPECT().IsClean("secret-token").Return(false)

	dec := JobOutputs(sm)
	task := &executor.JobTask{
		Run: func(context.Context) (*records.JobResult, error) {
			return result, nil
		},
	}

	task.Run = dec.DecorateJobRun(task)
	got, err := task.Run(context.Background())

	assert.NoError(t, err)
	assert.Same(t, result, got)
	assert.Equal(t, map[string]string{"clean": "public value"}, got.Outputs)
}
