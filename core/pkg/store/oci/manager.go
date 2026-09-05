/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ocistore

import (
	"context"

	"go.podman.io/storage"
)

// Manager manages OCI image storage and layer mounts for:
// - Immutable Actions
// - Pluggable runtimes
// - Sandboxer images
type Manager interface {
	Image(ctx context.Context, imageRef string) (*storage.Image, error)
	Pull(ctx context.Context, imageRef string, opts ...PullOption) (*storage.Image, error)
	Mount(ctx context.Context, image storage.Image, opts ...MountOption) (mountDir string, layerId string, err error)
	Unmount(ctx context.Context, layerId string) error
	Close() error
}
