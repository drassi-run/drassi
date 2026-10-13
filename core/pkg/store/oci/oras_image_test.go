/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ocistore

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
	oci "oras.land/oras-go/v2/content/oci"
)

func TestOrasImage_SpecAndRootFS(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	store, err := oci.New(dir)
	require.NoError(t, err)

	// Save dummy image config blob
	imgConfig := ocispec.Image{
		RootFS: ocispec.RootFS{
			Type: "layers",
			DiffIDs: []digest.Digest{
				digest.FromString("layer1"),
			},
		},
	}
	cfgBytes, err := json.Marshal(imgConfig)
	require.NoError(t, err)
	cfgDesc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageConfig,
		Digest:    digest.FromBytes(cfgBytes),
		Size:      int64(len(cfgBytes)),
	}
	err = store.Push(ctx, cfgDesc, bytes.NewReader(cfgBytes))
	require.NoError(t, err)

	// Save dummy manifest blob
	manifest := ocispec.Manifest{
		Config: cfgDesc,
		Layers: []ocispec.Descriptor{
			{
				MediaType: ocispec.MediaTypeImageLayerGzip,
				Digest:    digest.FromString("layer1"),
				Size:      10,
			},
		},
	}
	manifestBytes, err := json.Marshal(manifest)
	require.NoError(t, err)
	manifestDesc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageManifest,
		Digest:    digest.FromBytes(manifestBytes),
		Size:      int64(len(manifestBytes)),
	}
	err = store.Push(ctx, manifestDesc, bytes.NewReader(manifestBytes))
	require.NoError(t, err)

	err = store.Tag(ctx, manifestDesc, "example.com/test:latest")
	require.NoError(t, err)

	img := newOrasImage("example.com/test:latest", manifestDesc, store)
	require.Equal(t, "example.com/test:latest", img.Name())
	require.Equal(t, manifestDesc, img.Target())

	spec, err := img.Spec(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(spec.RootFS.DiffIDs))

	rootfs, err := img.RootFS(ctx)
	require.NoError(t, err)
	require.Equal(t, spec.RootFS.DiffIDs, rootfs)
}
