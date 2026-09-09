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
	"path/filepath"
	"strings"
	"sync"

	xfs "drassi.run/core/util/fs"
	xstring "drassi.run/core/util/string"
	"github.com/go-git/go-billy/v5/osfs"
	"go.podman.io/image/v5/copy"
	"go.podman.io/image/v5/docker"
	istorage "go.podman.io/image/v5/storage"
	"go.podman.io/image/v5/types"
	"go.podman.io/storage"
	"golang.org/x/sync/singleflight"
)

type Image = storage.Image
type ImageReference = types.ImageReference

// Manager manages OCI image storage and layer mounts for:
// - Immutable Actions
// - Pluggable runtimes
// - Sandboxer images
type Manager interface {
	Image(ctx context.Context, imageRef string) (*Image, error)
	Pull(ctx context.Context, imageRef string, opts ...PullOption) (*Image, error)
	Mount(ctx context.Context, image *Image, opts ...MountOption) (mountDir string, layerId string, err error)
	Unmount(ctx context.Context, layerId string) error
	Read(ctx context.Context, image *Image, opts ...ReadOption) (io.ReadCloser, error)
	Close() error
}

func New(store storage.Store) Manager {
	return &manager{store: store, pull: pull}
}

type manager struct {
	store storage.Store
	sf    singleflight.Group
	mu    sync.Mutex

	pull func(ctx context.Context, destRef, srcRef ImageReference, opts ...PullOption) error
}

func (m *manager) Image(_ context.Context, imageRef string) (*Image, error) {
	img, err := m.store.Image(imageRef)
	if err != nil {
		if errors.Is(err, storage.ErrImageUnknown) {
			return nil, nil
		}
		return nil, err
	}
	return img, nil
}

func (m *manager) Pull(ctx context.Context, imageRef string, opts ...PullOption) (*Image, error) {
	srcRef, err := docker.ParseReference(xstring.EnsurePrefix(imageRef, "//"))
	if err != nil {
		return nil, fmt.Errorf("parse source reference %q: %w", imageRef, err)
	}

	destRef, err := istorage.Transport.ParseStoreReference(m.store, imageRef)
	if err != nil {
		return nil, fmt.Errorf("parse destination reference %q: %w", imageRef, err)
	}

	v, err, _ := m.sf.Do(imageRef, func() (any, error) {
		if err := m.pull(ctx, destRef, srcRef, opts...); err != nil {
			return nil, fmt.Errorf("copy image %q: %w", imageRef, err)
		}

		if img, err := m.store.Image(imageRef); err != nil {
			return nil, fmt.Errorf("lookup pulled image %q: %w", imageRef, err)
		} else {
			return img, nil
		}
	})

	if err != nil {
		return nil, err
	}
	return v.(*Image), nil
}

func pull(ctx context.Context, destRef, srcRef ImageReference, opts ...PullOption) error {
	po := new(pullOptions)
	for _, opt := range opts {
		opt(po)
	}

	pc, ephemeral, err := po.PolicyCtx()
	if err != nil {
		return err
	}
	if ephemeral {
		defer pc.Destroy()
	}

	copyOpts := &copy.Options{
		ReportWriter:   po.ReportWriter,
		SourceCtx:      po.SystemContext,
		DestinationCtx: po.SystemContext,
	}

	_, err = copy.Image(ctx, pc, destRef, srcRef, copyOpts)
	return err
}

func (m *manager) Mount(_ context.Context, image *Image, opts ...MountOption) (string, string, error) {
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
func (m *manager) createCOWLayer(img *Image) (string, error) {
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

func (m *manager) Read(ctx context.Context, image *Image, opts ...ReadOption) (io.ReadCloser, error) {
	ro := new(readOptions)
	for _, opt := range opts {
		opt(ro)
	}

	layerId := image.TopLayer
	mountDir, err := m.store.Mount(layerId, "")
	if err != nil {
		return nil, fmt.Errorf("mount image: %w", err)
	}

	fsys := osfs.New(mountDir)
	subpath := strings.TrimPrefix(filepath.Clean(ro.Subpath), "/")

	if subpath != "" && subpath != "." {
		if info, err := fsys.Stat(subpath); err != nil {
			_, _ = m.store.Unmount(layerId, false)
			return nil, fmt.Errorf("resolve subpath %q: %w", ro.Subpath, err)
		} else if info.IsDir() {
			if fsys, err = fsys.Chroot(subpath); err != nil {
				_, _ = m.store.Unmount(layerId, false)
				return nil, fmt.Errorf("chroot subpath %q: %w", ro.Subpath, err)
			}
			subpath = "."
		}
	}

	rc := xfs.Read(ctx, fsys, subpath)
	rc = &mountReadCloser{
		ReadCloser: rc,
		unmount: func() {
			_, _ = m.store.Unmount(layerId, false)
		},
	}
	return rc, nil
}

type mountReadCloser struct {
	io.ReadCloser
	once    sync.Once
	unmount func()
}

func (m *mountReadCloser) Close() error {
	err := m.ReadCloser.Close()
	m.once.Do(m.unmount)
	return err
}

func (m *manager) Close() error {
	_, err := m.store.Shutdown(false)
	return err
}
