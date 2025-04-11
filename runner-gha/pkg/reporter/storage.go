package reporter

import (
	"context"
	"io"
)

type StoreManager interface {
	Get(kind string) Store
}

type Store interface {
	Put(ctx context.Context, url string, r io.Reader) error
}
