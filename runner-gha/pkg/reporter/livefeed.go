package reporter

import (
	"context"
	"encoding/json"
	"time"

	"drassi.run/core/util/context"
	"drassi.run/core/util/reactive"
	"github.com/chainguard-dev/clog"
	"github.com/coder/websocket"
)

func NewConsoleLiveFeeder(url string, contextual xcontext.Provider) (*ConsoleLiveFeeder, error) {
	ctx := contextual.Context()

	// TODO: header: Authorization, User-Agent
	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		return nil, err
	}

	batcher := reactive.NewThrottleBatcher[*line](100, 500*time.Millisecond)

	lf := &ConsoleLiveFeeder{
		conn:       conn,
		batcher:    batcher,
		contextual: contextual,
	}
	return lf, nil
}

type line struct {
	stepUid string
	number  int64
	content string
}

type liveFeed struct {
	StepUid   string   `json:"step_id"`
	Lines     []string `json:"value"`
	Count     int      `json:"count"`
	StartLine int64    `json:"start_line"`
}

type ConsoleLiveFeeder struct {
	conn       *websocket.Conn
	batcher    reactive.Batcher[*line]
	contextual xcontext.Provider
	logOffset  int64
}

func (lf *ConsoleLiveFeeder) Handle(_ context.Context, s string) error {
	l := &line{
		stepUid: "TODO", // TODO add stepUid
		number:  lf.logOffset,
		content: s,
	}
	lf.logOffset++

	return lf.batcher.Put(l)
}

func (lf *ConsoleLiveFeeder) Start() error {
	return lf.batcher.Start(lf.send)
}

func (lf *ConsoleLiveFeeder) Close() error {
	lf.batcher.Stop()
	return lf.conn.Close(websocket.StatusNormalClosure, "bye")
}

func (lf *ConsoleLiveFeeder) send(lines []*line) {
	ctx := lf.contextual.Context()

	var (
		stepUid string
		offset  int64
		msg     []string
	)

	// split lines into segments by stepUid
	var prev *line
	for _, curr := range lines {
		if prev != nil && prev.stepUid == curr.stepUid {
			msg = append(msg, curr.content)
			prev = curr
			continue
		}

		// curr is start of a new segment
		// => process the previous segment
		if err := lf.sendE(ctx, stepUid, msg, offset); err != nil {
			clog.Errorf("failed to upload logs: %v", err)
		}

		// save state of a new segment
		stepUid, offset = curr.stepUid, curr.number
		msg = []string{curr.content}
		prev = curr
	}

	// process the last segment
	if err := lf.sendE(ctx, stepUid, msg, offset); err != nil {
		clog.Errorf("failed to upload logs: %v", err)
	}

	return
}

func (lf *ConsoleLiveFeeder) sendE(ctx context.Context, stepUid string, lines []string, offset int64) error {
	if len(lines) == 0 {
		return nil
	}

	data := &liveFeed{
		StepUid:   stepUid,
		Lines:     lines,
		Count:     len(lines),
		StartLine: offset,
	}
	if w, err := lf.conn.Writer(ctx, websocket.MessageText); err != nil {
		return err
	} else if err = json.NewEncoder(w).Encode(data); err != nil {
		_ = w.Close()
		return err
	} else {
		return w.Close()
	}
}
