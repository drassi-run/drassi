/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ocistore

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	coreconfig "drassi.run/core/config"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
)

func TestOrasStore_EndToEnd(t *testing.T) {
	if _, err := exec.LookPath("fuse-overlayfs"); err != nil {
		t.Skip("fuse-overlayfs not installed, skipping integration mount test")
	}

	ctx := context.Background()
	storeDir := t.TempDir()
	mgr, err := NewOrasStore(&coreconfig.OrasStoreConfig{StoreDir: storeDir})
	require.NoError(t, err)
	defer mgr.Close()

	osStore := mgr.(*orasStore)

	// Create a test layer containing a test file
	var layerBuf bytes.Buffer
	gw := gzip.NewWriter(&layerBuf)
	tw := tar.NewWriter(gw)
	content := []byte("hello from oras test image")
	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name:     "hello.txt",
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     int64(len(content)),
	}))
	_, err = tw.Write(content)
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, gw.Close())

	layerBytes := layerBuf.Bytes()
	layerDigest := digest.FromBytes(layerBytes)
	diffID := digest.FromBytes([]byte("raw-uncompressed-diff"))
	layerDesc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageLayerGzip,
		Digest:    layerDigest,
		Size:      int64(len(layerBytes)),
	}
	require.NoError(t, osStore.store.Push(ctx, layerDesc, bytes.NewReader(layerBytes)))

	// Image Config
	imgConfig := ocispec.Image{
		RootFS: ocispec.RootFS{
			Type:    "layers",
			DiffIDs: []digest.Digest{diffID},
		},
	}
	cfgBytes, err := json.Marshal(imgConfig)
	require.NoError(t, err)
	cfgDesc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageConfig,
		Digest:    digest.FromBytes(cfgBytes),
		Size:      int64(len(cfgBytes)),
	}
	require.NoError(t, osStore.store.Push(ctx, cfgDesc, bytes.NewReader(cfgBytes)))

	// Manifest
	manifest := ocispec.Manifest{
		Config: cfgDesc,
		Layers: []ocispec.Descriptor{layerDesc},
	}
	manifestBytes, err := json.Marshal(manifest)
	require.NoError(t, err)
	manifestDesc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageManifest,
		Digest:    digest.FromBytes(manifestBytes),
		Size:      int64(len(manifestBytes)),
	}
	require.NoError(t, osStore.store.Push(ctx, manifestDesc, bytes.NewReader(manifestBytes)))
	require.NoError(t, osStore.store.Tag(ctx, manifestDesc, "test.local/img:v1"))

	// Pull with PullNever
	img, err := mgr.Pull(ctx, "test.local/img:v1", WithPullPolicy(PullNever))
	require.NoError(t, err)

	// Mount
	targetDir := filepath.Join(t.TempDir(), "target")
	mnt, err := mgr.Mount(ctx, img, WithTarget(targetDir), WithWritable(true))
	if err != nil && strings.Contains(err.Error(), "Function not implemented") {
		t.Skip("fuse mount not permitted in current environment")
	}
	require.NoError(t, err)
	defer mnt.Release(ctx)

	// Read
	rc, err := mgr.Read(ctx, img, WithSubpath("hello.txt"))
	require.NoError(t, err)
	defer rc.Close()

	tr := tar.NewReader(rc)
	hdr, err := tr.Next()
	require.NoError(t, err)
	require.Equal(t, "hello.txt", hdr.Name)

	readData, err := io.ReadAll(tr)
	require.NoError(t, err)
	require.Equal(t, content, readData)
}
