/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package xotel

import (
	"context"
	"log/slog"
	"sync"

	"github.com/chainguard-dev/clog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "drassi.run/core"

var once sync.Once
var tracer trace.Tracer
var meter metric.Meter

func setup() {
	tracer = otel.Tracer(instrumentationName)
	meter = otel.Meter(instrumentationName)
}

func T() trace.Tracer {
	once.Do(setup)
	return tracer
}

func M() metric.Meter {
	once.Do(setup)
	return meter
}

func StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return T().Start(ctx, spanName, opts...)
}

func EndSpan(span trace.Span, err *error) {
	if err != nil && *err != nil {
		span.RecordError(*err)
		span.SetStatus(codes.Error, (*err).Error())
	}
	span.End()
}

func ChildLogger(ctx context.Context, attrs []slog.Attr) (context.Context, *clog.Logger) {
	logger := clog.FromContext(ctx)
	if len(attrs) > 0 {
		args := make([]any, len(attrs))
		for i, attr := range attrs {
			args[i] = attr
		}

		logger = logger.With(args...)
		ctx = clog.WithLogger(ctx, logger)
	}

	return ctx, logger
}

func SetupTelemetry(ctx context.Context, method string, attrs ...attribute.KeyValue) (context.Context, func(*error)) {
	ctx, span := StartSpan(ctx, method, trace.WithAttributes(attrs...))
	ctx, logger := ChildLogger(ctx, ToSlogAttrs(attrs...))
	logger.Infof("start %s...", method)

	done := func(err *error) {
		if err != nil && *err != nil {
			logger.Errorf("end %s error: %v", method, *err)
		} else {
			logger.Infof("end %s", method)
		}
		EndSpan(span, err)
	}

	return ctx, done
}
