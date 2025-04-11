package store

import (
	"context"
	"io"
)

type Manager interface {
	Get(kind string) Store
}

type Store interface {
	Put(ctx context.Context, url string, r io.Reader) error
}
