/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package redact

import (
	"context"
	"io"
	"strings"
	"testing"

	mock_cmdtypes "drassi.run/core/mock/command/cmdtypes"
	mock_secret "drassi.run/core/mock/secret"
	"drassi.run/core/pkg/command/cmdtypes"
	"drassi.run/core/pkg/secret"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAttacherMasksAttachmentReader(t *testing.T) {
	ctrl := gomock.NewController(t)
	base := mock_cmdtypes.NewMockAttacher[string](ctrl)

	sm := secret.NewMasker()
	sm.AddSecret(secret.NewValueSecret("secret-token"))

	ctx := context.Background()
	att := &cmdtypes.Attachment{
		Type:   cmdtypes.STEP_SUMMARY,
		Reader: strings.NewReader("before secret-token after"),
	}

	base.EXPECT().
		Upload(ctx, "result-id", att).
		DoAndReturn(func(_ context.Context, _ string, got *cmdtypes.Attachment) error {
			body, err := io.ReadAll(got.Reader)
			require.NoError(t, err)
			assert.Equal(t, "before *** after", string(body))
			return nil
		})

	a := Attacher[string](base, sm)
	err := a.Upload(ctx, "result-id", att)
	assert.NoError(t, err)
}

func TestReporterMasksIssueMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	base := mock_cmdtypes.NewMockReporter[string](ctrl)

	sm := secret.NewMasker()
	sm.AddSecret(secret.NewValueSecret("secret-token"))

	ctx := context.Background()
	issue := &cmdtypes.Issue{
		Type:     cmdtypes.IssueTypeWarning,
		Category: "test",
		Message:  "do not leak secret-token",
		Data: map[string]string{
			"file": "main.go",
		},
	}

	base.EXPECT().
		AddIssue(ctx, "result-id", issue).
		DoAndReturn(func(_ context.Context, _ string, got *cmdtypes.Issue) error {
			assert.Equal(t, cmdtypes.IssueTypeWarning, got.Type)
			assert.Equal(t, "test", got.Category)
			assert.Equal(t, "do not leak ***", got.Message)
			assert.Equal(t, map[string]string{"file": "main.go"}, got.Data)
			return nil
		})

	r := Reporter[string](sm)(base)
	err := r.AddIssue(ctx, "result-id", issue)
	assert.NoError(t, err)
}

func TestReporterTruncatesMaskedIssueMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	base := mock_cmdtypes.NewMockReporter[string](ctrl)
	sm := mock_secret.NewMockMasker(ctrl)

	longMessage := strings.Repeat("x", issueMessageMaxLength+1)
	issue := &cmdtypes.Issue{Message: "raw message"}

	sm.EXPECT().Mask("raw message").Return(longMessage)
	base.EXPECT().
		AddIssue(gomock.Any(), "result-id", issue).
		DoAndReturn(func(_ context.Context, _ string, got *cmdtypes.Issue) error {
			assert.Len(t, got.Message, issueMessageMaxLength)
			assert.Equal(t, strings.Repeat("x", issueMessageMaxLength), got.Message)
			return nil
		})

	r := Reporter[string](sm)(base)
	err := r.AddIssue(context.Background(), "result-id", issue)
	assert.NoError(t, err)
}
