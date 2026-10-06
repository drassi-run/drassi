/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ocistore

import (
	"context"
	"errors"
	"fmt"
	"io"

	"drassi.run/core/config"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type Provider func(ctx context.Context) (Manager, error)

// Manager manages OCI image storage and rootfs mounts using containerd.
type Manager interface {
	// Pull fetches an image if needed (according to PullPolicy) and unpacks it into the snapshotter.
	Pull(ctx context.Context, imageRef string, opts ...PullOption) (Image, error)
	// Mount creates an active snapshot (view or COW) and mounts it to a host directory.
	Mount(ctx context.Context, image Image, opts ...MountOption) (*Mount, error)
	// Read opens a read-only stream to an image subpath or tar archive.
	Read(ctx context.Context, image Image, opts ...ReadOption) (io.ReadCloser, error)
	// Close closes the containerd client connection.
	Close() error
}

// Image represents an unpacked, locally available OCI image in the store.
// containerd.Image directly satisfies this interface.
type Image interface {
	Name() string
	Target() ocispec.Descriptor
	RootFS(ctx context.Context) ([]digest.Digest, error)
	Spec(ctx context.Context) (ocispec.Image, error)
}

// Mount represents an active filesystem mount of an image layer.
type Mount struct {
	// Target returns the directory path where the rootfs is mounted on the host.
	Target string
	// Release unmounts the filesystem and deletes the snapshot in containerd.
	// It is idempotent and safe for concurrent calls.
	Release func(ctx context.Context) error
}

// ErrImageNotFound is returned when an image cannot be found locally and PullPolicy is PullNever.
var ErrImageNotFound = errors.New("image not found")

// New creates an OCI image store Manager based on the provided configuration.
func New(cfg *config.OciStore) (Manager, error) {
	if cfg == nil {
		cfg = config.DefaultOciStoreConfig()
	}
	switch cfg.Backend {
	case "containerd", "":
		cCfg := cfg.Containerd
		if cCfg == nil {
			cCfg = config.DefaultContainerdStoreConfig()
		}
		return NewContainerdStore(cCfg)
	default:
		return nil, fmt.Errorf("unsupported oci store backend %q", cfg.Backend)
	}
}
