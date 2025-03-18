/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package storage

import (
	"context"
	"drassi.run/core/pkg/store/cache/types"
	"io"
)

type Storage interface {
	InitObject(ctx context.Context, cache *types.Cache) error
	WriteObject(ctx context.Context, cache *types.Cache, r io.Reader, offset, length int64) error
	CommitObject(ctx context.Context, cache *types.Cache) error
	ObjectLocation(ctx context.Context, cache *types.Cache) string
	ReadObject(ctx context.Context, cache *types.Cache, w io.Writer, offset, length int64) error
}
