/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package docker

import (
	"bytes"
	"context"
	"testing"
	"testing/fstest"

	"drassi.run/core/pkg/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageBuild_Validation(t *testing.T) {
	e := &engine{}
	ctx := context.Background()

	t.Run("neither context provided", func(t *testing.T) {
		err := e.ImageBuild(ctx, "test:tag", &container.BuildOptions{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exactly one of ContextTar or ContextFS")
	})

	t.Run("both contexts provided", func(t *testing.T) {
		err := e.ImageBuild(ctx, "test:tag", &container.BuildOptions{
			ContextTar: bytes.NewReader([]byte{}),
			ContextFS:  fstest.MapFS{},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exactly one of ContextTar or ContextFS")
	})
}
