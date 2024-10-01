package index

import (
	"context"
	"drassi.run/core/pkg/store/cache"
)

type Index interface {
	Create(ctx context.Context, cache *cache.Cache) error
	Update(ctx context.Context, cache *cache.Cache) error
	Get(ctx context.Context, id uint64) (*cache.Cache, error)
	Search(ctx context.Context, keys []string, version string) (*cache.Cache, error)
}
