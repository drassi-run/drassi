/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ocistore

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	coreconfig "drassi.run/core/config"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

type OrasStoreConfig = coreconfig.OrasStoreConfig

type orasStore struct {
	rootDir       string
	contentsDir   string
	layersDir     string
	snapshotsDir  string
	fuseOverlayfs string
	store         *oci.Store
}

func NewOrasStore(cfg *coreconfig.OrasStoreConfig) (Manager, error) {
	if cfg == nil {
		cfg = coreconfig.DefaultOrasStoreConfig()
	}
	storeDir := cfg.StoreDir
	if storeDir == "" {
		storeDir = filepath.Join(coreconfig.RootDir(), "ocistore")
	}

	contentsDir := filepath.Join(storeDir, "contents")
	if err := os.MkdirAll(contentsDir, 0755); err != nil {
		return nil, err
	}
	contentStore, err := oci.New(contentsDir)
	if err != nil {
		return nil, err
	}

	return &orasStore{
		rootDir:       storeDir,
		contentsDir:   contentsDir,
		layersDir:     filepath.Join(storeDir, "layers"),
		snapshotsDir:  filepath.Join(storeDir, "snapshots"),
		fuseOverlayfs: cfg.FuseOverlayfsBinary,
		store:         contentStore,
	}, nil
}

func (s *orasStore) Pull(ctx context.Context, imageRef string, opts ...PullOption) (Image, error) {
	po := &pullOptions{
		Policy: PullIfNotPresent,
	}
	for _, opt := range opts {
		opt(po)
	}

	if po.Policy == PullIfNotPresent || po.Policy == PullNever {
		desc, err := s.store.Resolve(ctx, imageRef)
		if err == nil {
			return newOrasImage(imageRef, desc, s.store), nil
		}
		if po.Policy == PullNever {
			return nil, ErrImageNotFound
		}
	}

	repo, err := remote.NewRepository(imageRef)
	if err != nil {
		return nil, fmt.Errorf("invalid image ref %q: %w", imageRef, err)
	}
	if po.InsecureSkipTLSVerify {
		repo.PlainHTTP = true
	}
	if po.AuthUsername != "" || po.AuthPassword != "" {
		repo.Client = &auth.Client{
			Credential: auth.StaticCredential(repo.Reference.Host(), auth.Credential{
				Username: po.AuthUsername,
				Password: po.AuthPassword,
			}),
		}
	}

	desc, err := oras.Copy(ctx, repo, imageRef, s.store, imageRef, oras.DefaultCopyOptions)
	if err != nil {
		return nil, fmt.Errorf("oras pull %q: %w", imageRef, err)
	}

	return newOrasImage(imageRef, desc, s.store), nil
}

func (s *orasStore) Mount(ctx context.Context, image Image, opts ...MountOption) (*Mount, error) {
	return s.mountFuseOverlayfs(ctx, image, opts...)
}

func (s *orasStore) Read(ctx context.Context, image Image, opts ...ReadOption) (io.ReadCloser, error) {
	return nil, nil
}

func (s *orasStore) Close() error {
	return nil
}
