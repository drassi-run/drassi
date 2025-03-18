/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package sandboxer

import (
	"context"
	"io"
	"io/fs"

	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/stream"
	"drassi.run/core/util/otel"
	"github.com/chainguard-dev/clog"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.opentelemetry.io/otel/trace"
)

func WithTelemetry(e Engine) Engine {
	if _, ok := e.(*telemetryEngine); ok {
		return e
	}
	return &telemetryEngine{Engine: e}
}

type telemetryEngine struct {
	Engine
}

func (e *telemetryEngine) Launch(ctx context.Context, req *LaunchRequest) (res *LaunchResponse, err error) {
	ctx, span := xotel.StartSpan(ctx, "Sandboxer.Launch")
	defer xotel.EndSpan(span, &err)

	clog.DebugContextf(ctx, "launching sandbox")

	if res, err = e.Engine.Launch(ctx, req); err != nil {
		return
	}

	s := withTelemetrySandbox(res.Sandbox)
	res.Sandbox = s

	if ce := res.ContainerEngine; ce != nil {
		ce = container.WithTelemetry(ce)
		res.ContainerEngine = ce
	}

	return
}

func withTelemetrySandbox(s Sandbox) Sandbox {
	if _, ok := s.(*telemetrySandbox); ok {
		return s
	}
	return &telemetrySandbox{Sandbox: s}
}

type telemetrySandbox struct {
	Sandbox
}

func (s *telemetrySandbox) Stat(ctx context.Context, path string) (fi fs.FileInfo, err error) {
	ctx, span := xotel.StartSpan(ctx, "Sandbox.Stat")
	defer xotel.EndSpan(span, &err)

	clog.DebugContextf(ctx, "stat path %q", path)

	return s.Sandbox.Stat(ctx, path)
}

func (s *telemetrySandbox) CopyIn(ctx context.Context, reader io.Reader, dst string) (err error) {
	ctx, span := xotel.StartSpan(ctx, "Sandbox.CopyIn",
		trace.WithAttributes(semconv.FilePath(dst)),
	)
	defer xotel.EndSpan(span, &err)

	clog.DebugContextf(ctx, "copy data into sandbox %q", dst)

	return s.Sandbox.CopyIn(ctx, reader, dst)
}

func (s *telemetrySandbox) CopyOut(ctx context.Context, src string) (r io.ReadCloser, err error) {
	ctx, span := xotel.StartSpan(ctx, "Sandbox.CopyOut",
		trace.WithAttributes(semconv.FilePath(src)),
	)
	defer xotel.EndSpan(span, &err)

	clog.DebugContextf(ctx, "copy data out sandbox %q", src)

	return s.Sandbox.CopyOut(ctx, src)
}

func (s *telemetrySandbox) Execute(ctx context.Context, cmd, path []string, env map[string]string, workdir string, streams stream.Streams) (err error) {
	ctx, span := xotel.StartSpan(ctx, "Sandbox.Execute")
	defer xotel.EndSpan(span, &err)

	clog.DebugContextf(ctx, "execute command in sandbox")

	return s.Sandbox.Execute(ctx, cmd, path, env, workdir, streams)
}
