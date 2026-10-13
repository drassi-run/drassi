/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ocistore

import (
	"bytes"
	"context"
	"testing"

	coreconfig "drassi.run/core/config"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
)

func TestOrasStore_PullPolicy(t *testing.T) {
	ctx := context.Background()
	storeDir := t.TempDir()
	mgr, err := NewOrasStore(&coreconfig.OrasStoreConfig{StoreDir: storeDir})
	require.NoError(t, err)

	// PullNever on non-existent image returns ErrImageNotFound
	_, err = mgr.Pull(ctx, "nonexistent.example.com/test:latest", WithPullPolicy(PullNever))
	require.ErrorIs(t, err, ErrImageNotFound)

	// Pre-populate an image into contents
	os := mgr.(*orasStore)
	manifestBytes := []byte(`{"schemaVersion":2,"layers":[]}`)
	desc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageManifest,
		Digest:    digest.FromBytes(manifestBytes),
		Size:      int64(len(manifestBytes)),
	}
	err = os.store.Push(ctx, desc, bytes.NewReader(manifestBytes))
	require.NoError(t, err)
	err = os.store.Tag(ctx, desc, "example.com/cached:v1")
	require.NoError(t, err)

	// PullIfNotPresent on existing image resolves from local store without network
	img, err := mgr.Pull(ctx, "example.com/cached:v1", WithPullPolicy(PullIfNotPresent))
	require.NoError(t, err)
	require.Equal(t, "example.com/cached:v1", img.Name())
	require.Equal(t, desc.Digest, img.Target().Digest)
}
