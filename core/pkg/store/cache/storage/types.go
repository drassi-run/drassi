package storage

import (
	"context"
	"io"

	"drassi.run/core/pkg/store/cache"
)

type Storage interface {
	InitObject(ctx context.Context, cache *cache.Cache) error
	WriteObject(ctx context.Context, cache *cache.Cache, r io.Reader, offset, length int64) error
	CommitObject(ctx context.Context, cache *cache.Cache) error
	ObjectLocation(ctx context.Context, cache *cache.Cache) string
	ReadObject(ctx context.Context, cache *cache.Cache, w io.Writer, offset, length int64) error
}
