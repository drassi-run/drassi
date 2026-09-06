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
	Mount(ctx context.Context, image *storage.Image, opts ...MountOption) (mountDir string, layerId string, err error)
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

func (m *manager) Mount(_ context.Context, image *storage.Image, opts ...MountOption) (string, string, error) {
	mo := new(mountOptions)
	for _, opt := range opts {
		opt(mo)
	}

	var err error
	layerId := image.TopLayer
	if mo.Writable {
		if layerId, err = m.createCOWLayer(image); err != nil {
			return "", "", err
		}
	}

	mountDir, err := m.store.Mount(layerId, "")
	if err != nil {
		if mo.Writable {
			_ = m.store.DeleteLayer(layerId)
		}
		return "", "", fmt.Errorf("mount image: %w", err)
	}
	return mountDir, layerId, nil
}

// Create writable layer on top of img.TopLayer
func (m *manager) createCOWLayer(img *storage.Image) (string, error) {
	layer, err := m.store.CreateLayer("", img.TopLayer, nil, "", true, nil)
	if err != nil {
		return "", fmt.Errorf("create COW layer: %w", err)
	}
	return layer.ID, nil
}

func (m *manager) Unmount(_ context.Context, layerId string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, err := m.store.Unmount(layerId, true)
	// If it's an ephemeral layer, delete it
	_ = m.store.DeleteLayer(layerId)
	return err
}

func (m *manager) Close() error {
	//TODO implement me
	panic("implement me")
}
