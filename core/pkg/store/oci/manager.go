/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ocistore

import (
	"context"
	"errors"
	"sync"

	"go.podman.io/storage"
	"golang.org/x/sync/singleflight"
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

func New(store storage.Store) Manager {
	return &manager{store: store}
}

type manager struct {
	store storage.Store
	sf    singleflight.Group
	mu    sync.Mutex
}

func (m *manager) Image(_ context.Context, imageRef string) (*storage.Image, error) {
	img, err := m.store.Image(imageRef)
	if err != nil {
		if errors.Is(err, storage.ErrImageUnknown) {
			return nil, nil
		}
		return nil, err
	}
	return img, nil
}

func (m *manager) Pull(ctx context.Context, imageRef string, opts ...PullOption) (*storage.Image, error) {
	//TODO implement me
	panic("implement me")
}

func (m *manager) Mount(ctx context.Context, image storage.Image, opts ...MountOption) (mountDir string, layerId string, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *manager) Unmount(ctx context.Context, layerId string) error {
	//TODO implement me
	panic("implement me")
}

func (m *manager) Close() error {
	//TODO implement me
	panic("implement me")
}
