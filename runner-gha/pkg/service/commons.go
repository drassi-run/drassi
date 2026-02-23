package service

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"drassi.run/core/pkg/stream"
	"drassi.run/core/util/http"
	"drassi.run/core/util/reactive"
	"drassi.run/gha-runner/pkg/types"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/streaming"
	"github.com/chainguard-dev/clog"
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

type LiveFeeder interface {
	stream.Handler
	io.Closer
	Start() error
}

type TimelineRecorder interface {
	Update(ctx context.Context, records ...*types.Record) error
}

type liveFeeder struct {
	SendFn    func(data *linesWrapper) error
	CloseFn   func() error
	batcher   reactive.Batcher[*line]
	logOffset int64
}

// https://github.com/actions/runner/blob/v2.324.0/src/Runner.Common/ResultsServer.cs#L220
func (lf *liveFeeder) Handle(_ context.Context, s string) error {
	l := &line{
		stepId:  "TODO", // TODO add stepId
		number:  lf.logOffset,
		content: s,
	}
	lf.logOffset++

	return lf.batcher.Put(l)
}

func (lf *liveFeeder) Start() error {
	return lf.batcher.Start(lf.send)
}

func (lf *liveFeeder) Close() error {
	lf.batcher.Stop()

	if lf.CloseFn != nil {
		return lf.CloseFn()
	}
	return nil
}

// https://github.com/actions/runner/blob/v2.324.0/src/Runner.Common/ResultsServer.cs#L220
func (lf *liveFeeder) send(lines []*line) {
	var (
		stepUid string
		offset  int64
		msg     []string
	)

	// split lines into segments by stepId
	var prev *line
	for _, curr := range lines {
		if prev != nil && prev.stepId == curr.stepId {
			msg = append(msg, curr.content)
			prev = curr
			continue
		}

		// curr is start of a new segment
		// => process the previous segment
		if err := lf.sendE(stepUid, msg, offset); err != nil {
			clog.Errorf("failed to upload logs: %v", err)
		}

		// save state of a new segment
		stepUid, offset = curr.stepId, curr.number
		msg = []string{curr.content}
		prev = curr
	}

	// process the last segment
	if err := lf.sendE(stepUid, msg, offset); err != nil {
		clog.Errorf("failed to upload logs: %v", err)
	}

	return
}

func (lf *liveFeeder) sendE(stepUid string, lines []string, offset int64) error {
	if len(lines) == 0 {
		return nil
	}

	data := &linesWrapper{
		Value:     lines,
		Count:     len(lines),
		StepId:    stepUid,
		StartLine: offset,
	}
	return lf.SendFn(data)
}

type SizeReaderAt interface {
	io.ReaderAt
	Size() int64
}

type rsc struct {
	io.ReadSeeker
	io.Closer
}

func reader(r io.Reader) (io.ReadSeekCloser, error) {
	if rs, ok := r.(io.ReadSeeker); ok {
		if c, ok := rs.(io.ReadSeekCloser); ok {
			return c, nil
		}
		return streaming.NopCloser(rs), nil
	}

	if sra, ok := r.(SizeReaderAt); ok {
		// SectionReader is ReadSeeker, but NOT Closer
		rs := io.NewSectionReader(sra, 0, sra.Size())
		if c, ok := r.(io.Closer); ok {
			return rsc{rs, c}, nil
		}
		return streaming.NopCloser(rs), nil
	}

	return nil, fmt.Errorf("unsupported reader type %T", r)
}
