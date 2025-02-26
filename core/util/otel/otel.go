package xotel

import (
	"context"
	"sync"

	"github.com/chainguard-dev/clog"
	"go.opentelemetry.io/otel"
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

func ChildLogger(ctx context.Context, args ...any) (context.Context, *clog.Logger) {
	logger := clog.FromContext(ctx)
	if len(args) > 0 {
		logger = logger.With(args...)
		ctx = clog.WithLogger(ctx, logger)
	}

	return ctx, logger
}
