package service

import (
	"context"
	"io"
	"net/http"

	"drassi.run/core/pkg/stream"
	"drassi.run/core/util/http"
	"drassi.run/gha-runner/pkg/types"
)

func newClient(url string, hc *http.Client) (*xhttp.Client, error) {
	client, err := xhttp.NewClient(url)
	if err != nil {
		return nil, err
	}

	client = client.WithDefaultErrorHandler(types.ParseActionsError).
		WithDefaultHeader("User-Agent", "gha-runner") // TODO

	if hc != nil {
		client = client.WithHttpClient(hc)
	}
	return client, nil
}

type Uploader interface {
	Upload(ctx context.Context, r io.Reader) error
	Complete(ctx context.Context, lineCount int64) error
}

type LiveFeeder interface {
	stream.Handler
	io.Closer
	Start() error
}
