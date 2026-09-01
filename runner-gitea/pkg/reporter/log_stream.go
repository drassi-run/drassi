/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package reporter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	"drassi.run/core/util/context"
	"drassi.run/core/util/reactive"
	runnerv1 "gitea.dev/actionslib/runner/v1"
	"gitea.dev/actionslib/runner/v1/runnerv1connect"
	"github.com/chainguard-dev/clog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewLogStreamer(
	taskId int64,
	contextual xcontext.Provider,
	client runnerv1connect.RunnerServiceClient,
) *LogStreamer {
	batcher := reactive.NewThrottleBatcher[*runnerv1.LogRow](50, 5*time.Second)
	ls := &LogStreamer{
		taskId:     taskId,
		contextual: contextual,
		batcher:    batcher,
		client:     client,
	}
	return ls
}

type LogStreamer struct {
	taskId    int64
	logOffset int64
	ackIndex  int64

	contextual xcontext.Provider
	batcher    reactive.Batcher[*runnerv1.LogRow]
	client     runnerv1connect.RunnerServiceClient
}

func (ls *LogStreamer) Start() error {
	return ls.batcher.Start(ls.uploadLog)
}

func (ls *LogStreamer) Close() error {
	ls.batcher.Stop()

	ctx := ls.contextual.Context()
	return ls.uploadLogE(ctx, nil, true)
}

func (ls *LogStreamer) uploadLog(logRows []*runnerv1.LogRow) {
	ctx := ls.contextual.Context()
	if err := ls.uploadLogE(ctx, logRows, false); err != nil {
		clog.Errorf("failed to upload logs: %v", err)
	}
}

// https://github.com/go-gitea/gitea/blob/v1.23.7/routers/api/actions/runner/runner.go#L238
func (ls *LogStreamer) uploadLogE(ctx context.Context, logRows []*runnerv1.LogRow, noMore bool) error {
	req := &runnerv1.UpdateLogRequest{
		TaskId: ls.taskId,
		Index:  ls.ackIndex,
		Rows:   logRows,
		NoMore: noMore,
	}
	resp, err := ls.client.UpdateLog(ctx, connect.NewRequest(req))
	if err != nil {
		return err
	}

	ack := resp.Msg.AckIndex
	if ack < ls.ackIndex+int64(len(logRows)) {
		return fmt.Errorf("submitted logs are lost")
	}
	ls.ackIndex = ack

	return nil
}

// ContextHandle is used for [scribe.Handler]
func (ls *LogStreamer) ContextHandle(_ context.Context, msg string) error {
	return ls.Handle(msg)
}

// Handle is used for [stream.Handler]
func (ls *LogStreamer) Handle(msg string) error {
	ls.logOffset++
	msg = strings.TrimRight(msg, "\r\n")

	row := &runnerv1.LogRow{
		Time:    timestamppb.Now(),
		Content: msg,
	}
	return ls.batcher.Put(row)
}

func (ls *LogStreamer) Offset() int64 {
	return ls.logOffset
}
