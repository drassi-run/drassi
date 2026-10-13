/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ocistore

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content/oci"
)

type orasImage struct {
	name  string
	desc  ocispec.Descriptor
	store *oci.Store
}

var _ Image = (*orasImage)(nil)

func newOrasImage(name string, desc ocispec.Descriptor, store *oci.Store) *orasImage {
	return &orasImage{
		name:  name,
		desc:  desc,
		store: store,
	}
}

func (img *orasImage) Name() string {
	return img.name
}

func (img *orasImage) Target() ocispec.Descriptor {
	return img.desc
}

func (img *orasImage) Manifest(ctx context.Context) (ocispec.Manifest, error) {
	rc, err := img.store.Fetch(ctx, img.desc)
	if err != nil {
		return ocispec.Manifest{}, fmt.Errorf("fetch manifest %s: %w", img.desc.Digest, err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return ocispec.Manifest{}, fmt.Errorf("read manifest: %w", err)
	}

	var m ocispec.Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return ocispec.Manifest{}, fmt.Errorf("unmarshal manifest: %w", err)
	}
	return m, nil
}

func (img *orasImage) Spec(ctx context.Context) (ocispec.Image, error) {
	m, err := img.Manifest(ctx)
	if err != nil {
		return ocispec.Image{}, err
	}

	rc, err := img.store.Fetch(ctx, m.Config)
	if err != nil {
		return ocispec.Image{}, fmt.Errorf("fetch image config %s: %w", m.Config.Digest, err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return ocispec.Image{}, fmt.Errorf("read image config: %w", err)
	}

	var cfg ocispec.Image
	if err := json.Unmarshal(data, &cfg); err != nil {
		return ocispec.Image{}, fmt.Errorf("unmarshal image config: %w", err)
	}
	return cfg, nil
}

func (img *orasImage) RootFS(ctx context.Context) ([]digest.Digest, error) {
	spec, err := img.Spec(ctx)
	if err != nil {
		return nil, err
	}
	return spec.RootFS.DiffIDs, nil
}
