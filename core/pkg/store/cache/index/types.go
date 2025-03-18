/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package index

import (
	"context"
	"drassi.run/core/pkg/store/cache/types"
)

type Index interface {
	Create(ctx context.Context, cache *types.Cache) error
	Update(ctx context.Context, cache *types.Cache) error
	Get(ctx context.Context, id uint64) (*types.Cache, error)
	Search(ctx context.Context, keys []string, version string) (*types.Cache, error)
}
