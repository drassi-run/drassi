package container

import (
	"context"
	"io"
	"io/fs"

	"drassi.run/core/pkg/container/types"
	"drassi.run/core/util/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.opentelemetry.io/otel/trace"
)

type telemetryEngine struct {
	Engine
}

func WithTelemetry(e Engine) Engine {
	if _, ok := e.(*telemetryEngine); !ok {
		return e
	}
	return &telemetryEngine{Engine: e}
}

func (e *telemetryEngine) ImagePull(ctx context.Context, ref string, opts *PullOptions) (err error) {
	ctx, span := xotel.StartSpan(ctx, "Container.ImagePull",
		trace.WithAttributes(semconv.ContainerImageName(ref)),
	)
	defer xotel.EndSpan(span, &err)

	return e.Engine.ImagePull(ctx, ref, opts)
}

func (e *telemetryEngine) ImageBuild(ctx context.Context, context io.Reader, opts *BuildOptions) (err error) {
	ctx, span := xotel.StartSpan(ctx, "Container.ImageBuild")
	defer xotel.EndSpan(span, &err)

	return e.Engine.ImageBuild(ctx, context, opts)
}

func (e *telemetryEngine) ContainerRun(ctx context.Context, spec *types.ContainerSpec, opts *RunOptions) (id string, err error) {
	ctx, span := xotel.StartSpan(ctx, "Container.ContainerRun")
	defer xotel.EndSpan(span, &err)

	if id, err = e.Engine.ContainerRun(ctx, spec, opts); err != nil {
		return
	}

	span.SetAttributes(semconv.ContainerID(id))
	return
}

func (e *telemetryEngine) ContainerExec(ctx context.Context, id string, opts *ExecOptions) (_ string, err error) {
	ctx, span := xotel.StartSpan(ctx, "Container.ContainerExec",
		trace.WithAttributes(semconv.ContainerID(id)),
	)
	defer xotel.EndSpan(span, &err)

	return e.Engine.ContainerExec(ctx, id, opts)
}

func (e *telemetryEngine) ContainerRemove(ctx context.Context, opts *RemoveOptions) (err error) {
	ctx, span := xotel.StartSpan(ctx, "Container.ContainerRemove")
	defer xotel.EndSpan(span, &err)

	return e.Engine.ContainerRemove(ctx, opts)
}

func (e *telemetryEngine) ContainerInspect(ctx context.Context, id string) (_ *types.ContainerSpec, err error) {
	ctx, span := xotel.StartSpan(ctx, "Container.ContainerInspect",
		trace.WithAttributes(semconv.ContainerID(id)),
	)
	defer xotel.EndSpan(span, &err)

	return e.Engine.ContainerInspect(ctx, id)
}

func (e *telemetryEngine) Stat(ctx context.Context, id string, path string) (_ fs.FileInfo, err error) {
	ctx, span := xotel.StartSpan(ctx, "Container.Stat",
		trace.WithAttributes(semconv.ContainerID(id), semconv.FilePath(path)),
	)
	defer xotel.EndSpan(span, &err)

	return e.Engine.Stat(ctx, id, path)
}

func (e *telemetryEngine) CopyIn(ctx context.Context, id string, opts *CopyInOptions) (err error) {
	ctx, span := xotel.StartSpan(ctx, "Container.CopyIn",
		trace.WithAttributes(semconv.ContainerID(id), semconv.FilePath(opts.DestinationPath)),
	)
	defer xotel.EndSpan(span, &err)

	return e.Engine.CopyIn(ctx, id, opts)
}

func (e *telemetryEngine) CopyOut(ctx context.Context, id string, opts *CopyOutOptions) (_ io.ReadCloser, err error) {
	ctx, span := xotel.StartSpan(ctx, "Container.CopyOut",
		trace.WithAttributes(semconv.ContainerID(id), semconv.FilePath(opts.SourcePath)),
	)
	defer xotel.EndSpan(span, &err)

	return e.Engine.CopyOut(ctx, id, opts)
}

func (e *telemetryEngine) NetworkCreate(ctx context.Context, spec *types.NetworkSpec) (_ string, err error) {
	ctx, span := xotel.StartSpan(ctx, "Container.NetworkCreate")
	defer xotel.EndSpan(span, &err)

	return e.Engine.NetworkCreate(ctx, spec)
}

func (e *telemetryEngine) NetworkRemove(ctx context.Context, opts *RemoveOptions) (err error) {
	ctx, span := xotel.StartSpan(ctx, "Container.NetworkRemove")
	defer xotel.EndSpan(span, &err)

	return e.Engine.NetworkRemove(ctx, opts)
}

func (e *telemetryEngine) VolumeCreate(ctx context.Context, spec *types.VolumeSpec) (_ string, err error) {
	ctx, span := xotel.StartSpan(ctx, "Container.VolumeCreate")
	defer xotel.EndSpan(span, &err)

	return e.Engine.VolumeCreate(ctx, spec)
}

func (e *telemetryEngine) VolumeRemove(ctx context.Context, opts *RemoveOptions) (err error) {
	ctx, span := xotel.StartSpan(ctx, "Container.VolumeRemove")
	defer xotel.EndSpan(span, &err)

	return e.Engine.VolumeRemove(ctx, opts)
}
