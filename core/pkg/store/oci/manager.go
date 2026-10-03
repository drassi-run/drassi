/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ocistore

import (
	"context"
	"io"
)

type Image = string // TODO

type Provider func(ctx context.Context) (Manager, error)

// Manager manages OCI image storage and layer mounts for:
// - Immutable Actions
// - Pluggable runtimes
// - Sandboxer images
// - Sandboxer providers
type Manager interface {
	Image(ctx context.Context, imageRef string) (*Image, error)
	Pull(ctx context.Context, imageRef string, opts ...PullOption) (*Image, error)
	Mount(ctx context.Context, image *Image, opts ...MountOption) (mountDir string, layerId string, err error)
	Unmount(ctx context.Context, layerId string) error
	Read(ctx context.Context, image *Image, opts ...ReadOption) (io.ReadCloser, error)
	Close() error
}
