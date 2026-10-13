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
	"os"
	"path/filepath"

	coreconfig "drassi.run/core/config"
	xfs "drassi.run/core/util/fs"
	"github.com/go-git/go-billy/v5/osfs"
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
	ro := &readOptions{}
	for _, opt := range opts {
		opt(ro)
	}

	tempDir, err := os.MkdirTemp("/tmp", "drassi-oci-read-*")
	if err != nil {
		return nil, fmt.Errorf("create temp mount dir for read: %w", err)
	}

	mnt, err := s.Mount(ctx, image, WithTarget(tempDir))
	if err != nil {
		_ = os.Remove(tempDir)
		return nil, fmt.Errorf("mount image for read: %w", err)
	}

	fsys := osfs.New(mnt.Target)
	stream := xfs.Read(ctx, fsys, ro.Subpath)

	r := &readCloser{
		Reader: stream,
		onClose: func() error {
			var errs []error
			if err := stream.Close(); err != nil {
				errs = append(errs, err)
			}
			if err := mnt.Release(context.Background()); err != nil {
				errs = append(errs, err)
			}
			if err := os.Remove(tempDir); err != nil && !os.IsNotExist(err) {
				errs = append(errs, err)
			}
			return errors.Join(errs...)
		},
	}
	return r, nil
}

func (s *orasStore) Close() error {
	return nil
}
