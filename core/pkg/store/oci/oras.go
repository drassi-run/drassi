/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ocistore

import (
	"context"
	"io"
	"path/filepath"

	coreconfig "drassi.run/core/config"
)

type OrasStoreConfig = coreconfig.OrasStoreConfig

type orasStore struct {
	rootDir       string
	contentsDir   string
	layersDir     string
	snapshotsDir  string
	fuseOverlayfs string
}

func NewOrasStore(cfg *coreconfig.OrasStoreConfig) (Manager, error) {
	if cfg == nil {
		cfg = coreconfig.DefaultOrasStoreConfig()
	}
	storeDir := cfg.StoreDir
	if storeDir == "" {
		storeDir = filepath.Join(coreconfig.RootDir(), "ocistore")
	}
	return &orasStore{
		rootDir:       storeDir,
		contentsDir:   filepath.Join(storeDir, "contents"),
		layersDir:     filepath.Join(storeDir, "layers"),
		snapshotsDir:  filepath.Join(storeDir, "snapshots"),
		fuseOverlayfs: cfg.FuseOverlayfsBinary,
	}, nil
}

func (s *orasStore) Pull(ctx context.Context, imageRef string, opts ...PullOption) (Image, error) {
	return nil, nil
}

func (s *orasStore) Mount(ctx context.Context, image Image, opts ...MountOption) (*Mount, error) {
	return nil, nil
}

func (s *orasStore) Read(ctx context.Context, image Image, opts ...ReadOption) (io.ReadCloser, error) {
	return nil, nil
}

func (s *orasStore) Close() error {
	return nil
}
