/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package firecracker

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamedDigest(t *testing.T) {
	digest := "sha256:" + strings.Repeat("a", 64)
	assert.Equal(t, digest, namedDigest("ghcr.io/drassi-run/ubuntu@"+digest))
	assert.Equal(t, digest, namedDigest("localhost:5000/ubuntu:26.04@"+digest))
	assert.Empty(t, namedDigest(defaultRootfsImage))
	assert.Empty(t, namedDigest("ghcr.io/drassi-run/ubuntu"))
}

func TestDigestCacheDir(t *testing.T) {
	assert.Equal(t, "sha256-"+strings.Repeat("a", 64),
		digestCacheDir("sha256:"+strings.Repeat("a", 64)))
}

func TestImageRefName(t *testing.T) {
	assert.Equal(t, "ghcr.io/drassi-run/ubuntu", imageRefName("ghcr.io/drassi-run/ubuntu:26.04"))
	assert.Equal(t, "ghcr.io/drassi-run/ubuntu", imageRefName("ghcr.io/drassi-run/ubuntu@sha256:abc"))
	assert.Equal(t, "localhost:5000/ubuntu", imageRefName("localhost:5000/ubuntu:26.04"))
	assert.Equal(t, "ubuntu", imageRefName("ubuntu"))
}

func TestDigestFromInspect(t *testing.T) {
	digest := "sha256:" + strings.Repeat("b", 64)
	other := "sha256:" + strings.Repeat("d", 64)
	id := "sha256:" + strings.Repeat("c", 64)
	img := dockerImageInspect{
		ID: id,
		RepoDigests: []string{
			"other.io/ubuntu@" + other,
			"ghcr.io/drassi-run/ubuntu@" + digest,
		},
	}
	assert.Equal(t, digest, digestFromInspect("ghcr.io/drassi-run/ubuntu:26.04", img))
	assert.Equal(t, other, digestFromInspect("ubuntu:26.04", dockerImageInspect{
		RepoDigests: []string{"other.io/ubuntu@" + other},
	}))
	assert.Equal(t, id, digestFromInspect("ubuntu:26.04", dockerImageInspect{ID: id}))
}

func TestMaterializeOCIRootfsUsesDigestCache(t *testing.T) {
	dir := t.TempDir()
	digest := "sha256:" + strings.Repeat("a", 64)
	out := filepath.Join(dir, "cache", digestCacheDir(digest), "rootfs.ext4")
	require.NoError(t, os.MkdirAll(filepath.Dir(out), 0o755))
	require.NoError(t, os.WriteFile(out, []byte("ext4"), 0o644))

	cfg := DefaultConfig()
	cfg.RootDir = dir
	cfg.RootfsPath = "oci://ghcr.io/drassi-run/ubuntu@" + digest

	path, err := cfg.materializeRootfs()
	require.NoError(t, err)
	assert.Equal(t, out, path)
}

func TestMaterializeFileRootfsUsesLocalExt4(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "rootfs.ext4")
	require.NoError(t, os.WriteFile(src, []byte("ext4"), 0o644))

	cfg := DefaultConfig()
	cfg.RootDir = dir
	cfg.RootfsPath = "file://" + src

	path, err := cfg.materializeRootfs()
	require.NoError(t, err)
	assert.Equal(t, src, path)
}

func TestParseRootfsRef(t *testing.T) {
	scheme, loc, err := parseRootfsRef("oci://ghcr.io/drassi-run/ubuntu:26.04")
	require.NoError(t, err)
	assert.Equal(t, "oci", scheme)
	assert.Equal(t, "ghcr.io/drassi-run/ubuntu:26.04", loc)

	scheme, loc, err = parseRootfsRef("file:///var/lib/rootfs.ext4")
	require.NoError(t, err)
	assert.Equal(t, "file", scheme)
	assert.Equal(t, "/var/lib/rootfs.ext4", loc)

	_, _, err = parseRootfsRef("https://example.com/rootfs.ext4")
	require.Error(t, err)
}

func TestParseKernelRef(t *testing.T) {
	scheme, loc, err := parseKernelRef("oci://ghcr.io/edera-dev/zone-kernel:6.18.43")
	require.NoError(t, err)
	assert.Equal(t, "oci", scheme)
	assert.Equal(t, "ghcr.io/edera-dev/zone-kernel:6.18.43", loc)

	scheme, loc, err = parseKernelRef("https://example.com/vmlinux")
	require.NoError(t, err)
	assert.Equal(t, "https", scheme)
	assert.Equal(t, "https://example.com/vmlinux", loc)

	scheme, loc, err = parseKernelRef("file:///boot/vmlinux")
	require.NoError(t, err)
	assert.Equal(t, "file", scheme)
	assert.Equal(t, "/boot/vmlinux", loc)

	scheme, loc, err = parseKernelRef("/boot/vmlinux")
	require.NoError(t, err)
	assert.Equal(t, "file", scheme)
	assert.Equal(t, "/boot/vmlinux", loc)

	_, _, err = parseKernelRef("ftp://example.com/vmlinux")
	require.Error(t, err)
}

func TestMaterializeOCIKernelUsesDigestCache(t *testing.T) {
	dir := t.TempDir()
	digest := "sha256:" + strings.Repeat("e", 64)
	out := filepath.Join(dir, "cache", digestCacheDir(digest), "kernel")
	require.NoError(t, os.MkdirAll(filepath.Dir(out), 0o755))
	require.NoError(t, os.WriteFile(out, append([]byte{0x7f, 'E', 'L', 'F'}, make([]byte, 16)...), 0o644))

	cfg := DefaultConfig()
	cfg.RootDir = dir
	cfg.KernelPath = "oci://ghcr.io/edera-dev/zone-kernel@" + digest

	path, err := cfg.materializeKernel()
	require.NoError(t, err)
	assert.Equal(t, out, path)
}

func TestMaterializeFileKernelUsesLocalELF(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "vmlinux")
	require.NoError(t, os.WriteFile(src, append([]byte{0x7f, 'E', 'L', 'F'}, make([]byte, 16)...), 0o644))

	cfg := DefaultConfig()
	cfg.RootDir = dir
	cfg.KernelPath = "file://" + src

	path, err := cfg.materializeKernel()
	require.NoError(t, err)
	assert.Equal(t, src, path)
}

func TestMaterializeHTTPKernelCachesByDigest(t *testing.T) {
	elf := append([]byte{0x7f, 'E', 'L', 'F'}, make([]byte, 16)...)
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = w.Write(elf)
	}))
	t.Cleanup(srv.Close)

	cfg := DefaultConfig()
	cfg.RootDir = t.TempDir()
	cfg.KernelPath = srv.URL + "/vmlinux"

	path, err := cfg.materializeKernel()
	require.NoError(t, err)
	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, elf, got)
	assert.Equal(t, 1, hits)

	path2, err := cfg.materializeKernel()
	require.NoError(t, err)
	assert.Equal(t, path, path2)
	assert.Equal(t, 1, hits)
}

func TestExtractELFKernelFromGzip(t *testing.T) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	elf := []byte{0x7f, 'E', 'L', 'F', 0, 1, 2, 3}
	_, err := w.Write(elf)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	out, err := extractELFKernel(append([]byte("MZ hdr"), buf.Bytes()...))
	require.NoError(t, err)
	assert.Equal(t, elf, out)

	_, err = extractELFKernel([]byte("MZ not a kernel"))
	require.Error(t, err)
}

func TestEnsureKernelRequiresPath(t *testing.T) {
	cfg := DefaultConfig()
	cfg.RootDir = t.TempDir()
	err := cfg.ensureKernel()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "kernel_path")
}

func TestMaterializeKernelFromOCI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping oci kernel extract in short mode")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not installed")
	}
	inspect := exec.Command("docker", "image", "inspect", "ghcr.io/edera-dev/zone-kernel:6.18.43")
	if err := inspect.Run(); err != nil {
		t.Skip("edera zone-kernel image not present")
	}

	cfg := DefaultConfig()
	cfg.RootDir = t.TempDir()
	cfg.KernelPath = "oci://ghcr.io/edera-dev/zone-kernel:6.18.43"
	require.NoError(t, cfg.ensureKernel())
	st, err := os.Stat(cfg.kernel)
	require.NoError(t, err)
	assert.Greater(t, st.Size(), int64(1<<20))
}

func TestDefaultKernelArgsUseTiniInit(t *testing.T) {
	args := DefaultConfig().KernelArgs
	assert.Contains(t, args, "init="+guestTiniPath+" -- "+guestInitPath)
	assert.NotContains(t, args, guestTiniPath+" -- "+guestAgentPath)
}

func TestInstallGuestInit(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, installGuestInit(dir))

	path := filepath.Join(dir, "usr", "sbin", "drassi-init")
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o755), info.Mode().Perm())

	body, err := os.ReadFile(path)
	require.NoError(t, err)
	got := string(body)
	assert.True(t, strings.HasPrefix(got, "#!/bin/sh"))
	assert.Contains(t, got, "dockerd")
	assert.Contains(t, got, "exec "+guestAgentPath)
}

func TestEnsureRootfsRequiresPath(t *testing.T) {
	cfg := DefaultConfig()
	cfg.RootDir = t.TempDir()
	cfg.RootfsPath = ""
	err := cfg.ensureRootfs()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rootfs_path")
}
