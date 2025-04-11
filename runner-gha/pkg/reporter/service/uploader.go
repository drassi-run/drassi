package service

import (
	"context"
	"io"
)

type Uploader interface {
	Upload(ctx context.Context, r io.Reader) error
	Complete(ctx context.Context, lineCount int64) error
}
