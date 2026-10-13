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
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	xerror "drassi.run/core/util/error"
	xsync "drassi.run/core/util/sync"
	"github.com/google/uuid"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func buildFuseOverlayArgs(lowers []string, upper, work, target string) []string {
	var opts []string
	if len(lowers) > 0 {
		opts = append(opts, "lowerdir="+strings.Join(lowers, ":"))
	}
	if upper != "" {
		opts = append(opts, "upperdir="+upper)
	}
	if work != "" {
		opts = append(opts, "workdir="+work)
	}

	args := []string{}
	if len(opts) > 0 {
		args = append(args, "-o", strings.Join(opts, ","))
	}
	args = append(args, target)
	return args
}

func unmountFUSE(ctx context.Context, target string) error {
	// Try fusermount3 first, fallback to fusermount
	var errs []error
	if fuse3, err := exec.LookPath("fusermount3"); err == nil {
		cmd := exec.CommandContext(ctx, fuse3, "-u", target)
		if out, err := cmd.CombinedOutput(); err == nil {
			return nil
		} else {
			errs = append(errs, fmt.Errorf("fusermount3 -u %s: %w (output: %s)", target, err, string(out)))
		}
	}

	if fuse, err := exec.LookPath("fusermount"); err == nil {
		cmd := exec.CommandContext(ctx, fuse, "-u", target)
		if out, err := cmd.CombinedOutput(); err == nil {
			return nil
		} else {
			errs = append(errs, fmt.Errorf("fusermount -u %s: %w (output: %s)", target, err, string(out)))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return errors.New("neither fusermount3 nor fusermount found in PATH")
}

func (s *orasStore) resolveFuseOverlayfs() (string, error) {
	if s.fuseOverlayfs != "" {
		return s.fuseOverlayfs, nil
	}
	bin, err := exec.LookPath("fuse-overlayfs")
	if err != nil {
		return "", fmt.Errorf("fuse-overlayfs binary not found: %w", err)
	}
	return bin, nil
}

func (s *orasStore) ensureLayerUnpacked(ctx context.Context, desc ocispec.Descriptor, diffID digest.Digest) (string, error) {
	diffIDStr := diffID.Encoded()
	layerDir := filepath.Join(s.layersDir, diffIDStr)

	if fi, err := os.Stat(layerDir); err == nil && fi.IsDir() {
		return layerDir, nil
	}

	if err := os.MkdirAll(s.layersDir, 0755); err != nil {
		return "", fmt.Errorf("create layers dir: %w", err)
	}

	// Atomic unpack using temporary directory
	tmpDir, err := os.MkdirTemp(s.layersDir, ".tmp-*")
	if err != nil {
		return "", fmt.Errorf("create temp layer dir: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	rc, err := s.store.Fetch(ctx, desc)
	if err != nil {
		return "", fmt.Errorf("fetch layer blob %s: %w", desc.Digest, err)
	}
	defer rc.Close()

	if err := extractLayer(ctx, rc, tmpDir); err != nil {
		return "", fmt.Errorf("extract layer %s (diffID %s): %w", desc.Digest, diffID, err)
	}

	if err := os.Rename(tmpDir, layerDir); err != nil {
		// If another goroutine unpacked concurrently, check if target dir exists
		if fi, statErr := os.Stat(layerDir); statErr == nil && fi.IsDir() {
			return layerDir, nil
		}
		return "", fmt.Errorf("rename unpacked layer to %s: %w", layerDir, err)
	}

	return layerDir, nil
}

func (s *orasStore) mountFuseOverlayfs(ctx context.Context, image Image, opts ...MountOption) (*Mount, error) {
	mo := &mountOptions{}
	for _, opt := range opts {
		opt(mo)
	}

	if mo.Block {
		return nil, errors.New("block device mount not supported by oras store")
	}

	fuseBin, err := s.resolveFuseOverlayfs()
	if err != nil {
		return nil, err
	}

	orasImg, ok := image.(*orasImage)
	if !ok {
		return nil, fmt.Errorf("unsupported image type %T for oras store", image)
	}

	manifest, err := orasImg.Manifest(ctx)
	if err != nil {
		return nil, fmt.Errorf("get manifest: %w", err)
	}

	diffIDs, err := orasImg.RootFS(ctx)
	if err != nil {
		return nil, fmt.Errorf("get rootfs diffIDs: %w", err)
	}

	if len(diffIDs) != len(manifest.Layers) {
		return nil, fmt.Errorf("mismatch diffIDs count (%d) and layers count (%d)", len(diffIDs), len(manifest.Layers))
	}

	// Unpack layers and collect directory paths
	layerDirs := make([]string, len(diffIDs))
	for i, diffID := range diffIDs {
		dir, err := s.ensureLayerUnpacked(ctx, manifest.Layers[i], diffID)
		if err != nil {
			return nil, err
		}
		layerDirs[i] = dir
	}

	// In overlayfs, lowerdir order is top-most to bottom-most: diffID_N:...:diffID_0
	lowers := make([]string, len(layerDirs))
	for i := range layerDirs {
		lowers[i] = layerDirs[len(layerDirs)-1-i]
	}

	target := mo.Target
	if target == "" {
		return nil, errors.New("mount target directory is required")
	}
	if err := os.MkdirAll(target, 0755); err != nil {
		return nil, fmt.Errorf("create target dir: %w", err)
	}

	var rb xerror.Rollbacker
	defer rb.Run(ctx)

	var upperDir, workDir, snapshotDir string
	if mo.Writable {
		snapID := uuid.NewString()
		snapshotDir = filepath.Join(s.snapshotsDir, snapID)
		upperDir = filepath.Join(snapshotDir, "upper")
		workDir = filepath.Join(snapshotDir, "work")

		if err := os.MkdirAll(upperDir, 0755); err != nil {
			return nil, fmt.Errorf("create snapshot upper dir: %w", err)
		}
		if err := os.MkdirAll(workDir, 0755); err != nil {
			return nil, fmt.Errorf("create snapshot work dir: %w", err)
		}

		rb.Add(func(ctx context.Context) error {
			return os.RemoveAll(snapshotDir)
		})
	}

	args := buildFuseOverlayArgs(lowers, upperDir, workDir, target)
	cmd := exec.CommandContext(ctx, fuseBin, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("run fuse-overlayfs: %w (output: %s)", err, string(out))
	}

	rb.Add(func(ctx context.Context) error {
		return unmountFUSE(ctx, target)
	})

	rb.Dismiss()

	mnt := &Mount{
		Target:  target,
		Release: xsync.InvokeOnce(rb.Do),
	}
	return mnt, nil
}
