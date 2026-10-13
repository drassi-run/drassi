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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	coreconfig "drassi.run/core/config"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
)

func TestBuildFuseOverlayArgs(t *testing.T) {
	lowers := []string{"/layer2", "/layer1"}
	target := "/mnt/target"

	// Read-only
	argsRO := buildFuseOverlayArgs(lowers, "", "", target)
	require.Equal(t, []string{"-o", "lowerdir=/layer2:/layer1", target}, argsRO)

	// Writable
	argsRW := buildFuseOverlayArgs(lowers, "/upper", "/work", target)
	require.Equal(t, []string{"-o", "lowerdir=/layer2:/layer1,upperdir=/upper,workdir=/work", target}, argsRW)
}

func TestOrasStore_Mount_Validation(t *testing.T) {
	ctx := context.Background()
	storeDir := t.TempDir()
	mgr, err := NewOrasStore(&coreconfig.OrasStoreConfig{
		StoreDir:            storeDir,
		FuseOverlayfsBinary: "/nonexistent/fuse-overlayfs",
	})
	require.NoError(t, err)

	osStore := mgr.(*orasStore)

	// Block option not supported
	_, err = osStore.Mount(ctx, nil, WithBlock())
	require.ErrorContains(t, err, "block device mount not supported")

	// Missing fuse-overlayfs
	_, err = NewOrasStore(&coreconfig.OrasStoreConfig{
		StoreDir:            storeDir,
		FuseOverlayfsBinary: "/nonexistent/bin",
	})
	require.NoError(t, err)
}

func TestOrasStore_EnsureLayerUnpacked(t *testing.T) {
	ctx := context.Background()
	storeDir := t.TempDir()
	mgr, err := NewOrasStore(&coreconfig.OrasStoreConfig{StoreDir: storeDir})
	require.NoError(t, err)
	osStore := mgr.(*orasStore)

	// Prepare tar.gz layer content
	var layerBuf bytes.Buffer
	gw := gzip.NewWriter(&layerBuf)
	tw := tar.NewWriter(gw)
	fileData := []byte("hello layer unpacking")
	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name:     "app/config.json",
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     int64(len(fileData)),
	}))
	_, err = tw.Write(fileData)
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, gw.Close())

	layerBytes := layerBuf.Bytes()
	layerDigest := digest.FromBytes(layerBytes)
	diffID := digest.FromString("test-diff-id-1")

	desc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageLayerGzip,
		Digest:    layerDigest,
		Size:      int64(len(layerBytes)),
	}

	err = osStore.store.Push(ctx, desc, bytes.NewReader(layerBytes))
	require.NoError(t, err)

	// Unpack layer
	unpackedDir, err := osStore.ensureLayerUnpacked(ctx, desc, diffID)
	require.NoError(t, err)
	require.DirExists(t, unpackedDir)

	content, err := os.ReadFile(filepath.Join(unpackedDir, "app", "config.json"))
	require.NoError(t, err)
	require.Equal(t, fileData, content)

	// Calling again should be cached (idempotent)
	cachedDir, err := osStore.ensureLayerUnpacked(ctx, desc, diffID)
	require.NoError(t, err)
	require.Equal(t, unpackedDir, cachedDir)
}

func TestOrasStore_MountAndRelease(t *testing.T) {
	if _, err := exec.LookPath("fuse-overlayfs"); err != nil {
		t.Skip("fuse-overlayfs not available, skipping mount test")
	}

	ctx := context.Background()
	storeDir := t.TempDir()
	mgr, err := NewOrasStore(&coreconfig.OrasStoreConfig{StoreDir: storeDir})
	require.NoError(t, err)
	osStore := mgr.(*orasStore)

	// Layer 1
	var l1Buf bytes.Buffer
	gw1 := gzip.NewWriter(&l1Buf)
	tw1 := tar.NewWriter(gw1)
	l1Content := []byte("bottom layer")
	require.NoError(t, tw1.WriteHeader(&tar.Header{
		Name:     "base.txt",
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     int64(len(l1Content)),
	}))
	_, err = tw1.Write(l1Content)
	require.NoError(t, err)
	require.NoError(t, tw1.Close())
	require.NoError(t, gw1.Close())

	l1Bytes := l1Buf.Bytes()
	l1Desc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageLayerGzip,
		Digest:    digest.FromBytes(l1Bytes),
		Size:      int64(len(l1Bytes)),
	}
	require.NoError(t, osStore.store.Push(ctx, l1Desc, bytes.NewReader(l1Bytes)))

	// Image config
	l1DiffID := digest.FromString("l1-diff")
	imgConfig := ocispec.Image{
		RootFS: ocispec.RootFS{
			Type:    "layers",
			DiffIDs: []digest.Digest{l1DiffID},
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
		Layers: []ocispec.Descriptor{l1Desc},
	}
	manifestBytes, err := json.Marshal(manifest)
	require.NoError(t, err)
	manifestDesc := ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageManifest,
		Digest:    digest.FromBytes(manifestBytes),
		Size:      int64(len(manifestBytes)),
	}
	require.NoError(t, osStore.store.Push(ctx, manifestDesc, bytes.NewReader(manifestBytes)))

	img := newOrasImage("test:v1", manifestDesc, osStore.store)

	targetDir := filepath.Join(t.TempDir(), "mnt")
	mnt, err := mgr.Mount(ctx, img, WithTarget(targetDir), WithWritable(true))
	if err != nil && strings.Contains(err.Error(), "Function not implemented") {
		t.Skip("fuse mount not permitted in current environment")
	}
	require.NoError(t, err)
	require.NotNil(t, mnt)
	require.Equal(t, targetDir, mnt.Target)

	// Verify content is readable from mount
	readBytes, err := os.ReadFile(filepath.Join(targetDir, "base.txt"))
	require.NoError(t, err)
	require.Equal(t, l1Content, readBytes)

	// Release mount
	err = mnt.Release(ctx)
	require.NoError(t, err)

	// Release again should be idempotent (no error)
	err = mnt.Release(ctx)
	require.NoError(t, err)
}


