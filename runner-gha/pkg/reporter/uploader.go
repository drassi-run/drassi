package reporter

import (
	"context"
	"io"
)

type Uploader interface {
	Upload(ctx context.Context, url string, r io.Reader) error
}
