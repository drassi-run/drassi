# Oras OCI Store Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement `ocistore.Manager` using the `oras-go` library and `fuse-overlayfs` for rootless, daemonless OCI image storage and mounting.

**Architecture:** Use `oras.land/oras-go/v2/content/oci` to store OCI image manifests and blobs in `<RootDir>/ocistore/contents`. Unpack layer archives into `<RootDir>/ocistore/layers/<diffID>` on demand. Mount layered rootfs using `fuse-overlayfs` with `<RootDir>/ocistore/snapshots/<uuid>` upper/work dirs for writable mounts, releasing mounts cleanly via `fusermount3 -u` / `fusermount -u` without root privileges.

**Tech Stack:** Go 1.27, `oras.land/oras-go/v2`, `fuse-overlayfs`, `fusermount3`/`fusermount`, OCI Image Spec (`github.com/opencontainers/image-spec`).

**Spec:** `docs/superpowers/specs/2026-10-13-oras-ocistore-design.md`

## Global Constraints

- SPDX license header `/*\n * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors\n *\n * SPDX-License-Identifier: Apache-2.0\n */` required on all newly created `.go` files.
- Store directory default: `filepath.Join(coreconfig.RootDir(), "ocistore")`.
- Rootless operations: no commands or permissions requiring `sudo` or `root`.
- Backend config name in `core/config/config.go`: `"oras"`.

---

### Task 1: Configuration Support for Oras Store

**Files:**
- Modify: `core/config/config.go:33-43`
- Modify: `core/config/defaults.go:48-53`
- Modify: `core/pkg/store/oci/manager.go:56-70`
- Test: `core/config/config_test.go`
- Test: `core/pkg/store/oci/manager_test.go`

**Interfaces:**
- Consumes: `core/config.RootDir()`
- Produces:
  ```go
  type OrasStoreConfig struct {
      StoreDir            string `toml:"store_dir,omitempty" json:"store_dir,omitempty"`
      FuseOverlayfsBinary string `toml:"fuse_overlayfs_binary,omitempty" json:"fuse_overlayfs_binary,omitempty"`
  }
  ```

- [ ] **Step 1: Write the failing test for configuration and store instantiation**

Create `core/pkg/store/oci/manager_test.go`:
```go
/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ocistore_test

import (
	"testing"

	"drassi.run/core/config"
	ocistore "drassi.run/core/pkg/store/oci"
	"github.com/stretchr/testify/require"
)

func TestNew_OrasBackend(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.OciStore{
		Backend: "oras",
		Oras: &config.OrasStoreConfig{
			StoreDir: tempDir,
		},
	}
	mgr, err := ocistore.New(cfg)
	require.NoError(t, err)
	require.NotNil(t, mgr)
	require.NoError(t, mgr.Close())
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `(cd core && go test -v ./pkg/store/oci -run TestNew_OrasBackend)`
Expected: FAIL with "unsupported oci store backend \"oras\"" or compilation error.

- [ ] **Step 3: Update config structs and New dispatcher**

In `core/config/config.go`:
```go
type OciStore struct {
	Backend    string                 `toml:"backend" json:"backend"`
	Containerd *ContainerdStoreConfig `toml:"containerd,omitempty" json:"containerd,omitempty"`
	Oras       *OrasStoreConfig       `toml:"oras,omitempty" json:"oras,omitempty"`
}

type OrasStoreConfig struct {
	StoreDir            string `toml:"store_dir,omitempty" json:"store_dir,omitempty"`
	FuseOverlayfsBinary string `toml:"fuse_overlayfs_binary,omitempty" json:"fuse_overlayfs_binary,omitempty"`
}
```

In `core/pkg/store/oci/manager.go`:
```go
	case "oras":
		oCfg := cfg.Oras
		if oCfg == nil {
			oCfg = &config.OrasStoreConfig{}
		}
		return NewOrasStore(oCfg)
```

Create stub `NewOrasStore` in `core/pkg/store/oci/oras.go`:
```go
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
	rootDir         string
	contentsDir     string
	layersDir       string
	snapshotsDir    string
	fuseOverlayfs   string
}

func NewOrasStore(cfg *coreconfig.OrasStoreConfig) (Manager, error) {
	storeDir := cfg.StoreDir
	if storeDir == "" {
		storeDir = filepath.Join(coreconfig.RootDir(), "ocistore")
	}
	return &orasStore{
		rootDir:      storeDir,
		contentsDir:  filepath.Join(storeDir, "contents"),
		layersDir:    filepath.Join(storeDir, "layers"),
		snapshotsDir: filepath.Join(storeDir, "snapshots"),
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `(cd core && go test -v ./pkg/store/oci -run TestNew_OrasBackend)`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add core/config/config.go core/pkg/store/oci/manager.go core/pkg/store/oci/oras.go core/pkg/store/oci/manager_test.go
git commit -m "feat(ocistore): add oras backend configuration and factory dispatch"
```

---

### Task 2: Implement Oras Image Model and Local Content Storage

**Files:**
- Create: `core/pkg/store/oci/oras_image.go`
- Modify: `core/pkg/store/oci/oras.go`
- Test: `core/pkg/store/oci/oras_image_test.go`

**Interfaces:**
- Consumes: `oras.land/oras-go/v2/content/oci`, `github.com/opencontainers/image-spec`
- Produces: `orasImage` implementing `ocistore.Image`:
  - `Name() string`
  - `Target() ocispec.Descriptor`
  - `RootFS(ctx context.Context) ([]digest.Digest, error)`
  - `Spec(ctx context.Context) (ocispec.Image, error)`

- [ ] **Step 1: Write the failing test for `orasImage`**

Create `core/pkg/store/oci/oras_image_test.go`:
```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `(cd core && go test -v ./pkg/store/oci -run TestOrasImage_SpecAndRootFS)`
Expected: FAIL with undefined `newOrasImage`.

- [ ] **Step 3: Implement `orasImage`**

Create `core/pkg/store/oci/oras_image.go`:
```go
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
	name   string
	desc   ocispec.Descriptor
	store  *oci.Store
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `(cd core && go test -v ./pkg/store/oci -run TestOrasImage_SpecAndRootFS)`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add core/pkg/store/oci/oras_image.go core/pkg/store/oci/oras_image_test.go
git commit -m "feat(ocistore): implement orasImage struct conforming to ocistore.Image"
```

---

### Task 3: Implement Pull Workflow with Policy and Credentials

**Files:**
- Modify: `core/pkg/store/oci/oras.go`
- Test: `core/pkg/store/oci/oras_pull_test.go`

**Interfaces:**
- Consumes: `oras.Copy`, `oras.land/oras-go/v2/registry/remote`, `PullOption`
- Produces: `(s *orasStore) Pull(ctx context.Context, imageRef string, opts ...PullOption) (Image, error)`

- [ ] **Step 1: Write the failing test for Pull policy handling (PullNever and local resolution)**

Create `core/pkg/store/oci/oras_pull_test.go`:
```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `(cd core && go test -v ./pkg/store/oci -run TestOrasStore_PullPolicy)`
Expected: FAIL

- [ ] **Step 3: Implement Pull method in `core/pkg/store/oci/oras.go`**

Implement initialization of `os.store` (`oci.New(contentsDir)`), local tag resolution, remote repository initialization with auth and TLS options, and `oras.Copy`:
```go
// In NewOrasStore:
if err := os.MkdirAll(storeDir, 0755); err != nil {
	return nil, err
}
contentStore, err := oci.New(contentsDir)
if err != nil {
	return nil, err
}
// In Pull:
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `(cd core && go test -v ./pkg/store/oci -run TestOrasStore_PullPolicy)`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add core/pkg/store/oci/oras.go core/pkg/store/oci/oras_pull_test.go
git commit -m "feat(ocistore): implement oras pull logic with policy and credentials"
```

---

### Task 4: Implement Layer Archive Extraction and Cache

**Files:**
- Create: `core/pkg/store/oci/extractor.go`
- Modify: `core/pkg/store/oci/oras.go`
- Test: `core/pkg/store/oci/extractor_test.go`

**Interfaces:**
- Consumes: `io.Reader`, tar stream
- Produces: `extractLayer(ctx context.Context, r io.Reader, targetDir string) error`

- [ ] **Step 1: Write the failing test for layer archive extraction**

Create `core/pkg/store/oci/extractor_test.go`:
```go
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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractLayer_Basic(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	// Add directory
	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name:     "etc/",
		Typeflag: tar.TypeDir,
		Mode:     0755,
	}))

	// Add file
	content := []byte("hello oras")
	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name:     "etc/test.conf",
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     int64(len(content)),
	}))
	_, err := tw.Write(content)
	require.NoError(t, err)

	require.NoError(t, tw.Close())
	require.NoError(t, gw.Close())

	destDir := t.TempDir()
	err = extractLayer(context.Background(), &buf, destDir)
	require.NoError(t, err)

	readContent, err := os.ReadFile(filepath.Join(destDir, "etc", "test.conf"))
	require.NoError(t, err)
	require.Equal(t, content, readContent)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `(cd core && go test -v ./pkg/store/oci -run TestExtractLayer_Basic)`
Expected: FAIL with undefined `extractLayer`.

- [ ] **Step 3: Implement `extractLayer` and layer unpack synchronization**

Create `core/pkg/store/oci/extractor.go`:
Handle gzip/zstd detection, tar file unpacking with security bounds (`filepath.Clean`, preventing path traversal), whiteout marker handling (`.wh.<name>`), and directory creation:
```go
/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ocistore

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"drassi.run/core/util/compression"
)

func extractLayer(ctx context.Context, r io.Reader, targetDir string) error {
	xr, err := compression.DecompressStream(r)
	if err != nil {
		return fmt.Errorf("decompress layer stream: %w", err)
	}
	defer xr.Close()

	tr := tar.NewReader(xr)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar entry: %w", err)
		}

		cleanPath := filepath.Clean(hdr.Name)
		if strings.HasPrefix(cleanPath, "..") || filepath.IsAbs(cleanPath) {
			continue // skip dangerous path
		}

		destPath := filepath.Join(targetDir, cleanPath)
		baseName := filepath.Base(cleanPath)
		dirName := filepath.Dir(destPath)

		// Handle whiteouts
		if strings.HasPrefix(baseName, ".wh.") {
			targetName := strings.TrimPrefix(baseName, ".wh.")
			_ = os.RemoveAll(filepath.Join(dirName, targetName))
			continue
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(destPath, hdr.FileInfo().Mode()); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(dirName, 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(destPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, hdr.FileInfo().Mode())
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		case tar.TypeSymlink:
			if err := os.MkdirAll(dirName, 0755); err != nil {
				return err
			}
			_ = os.Remove(destPath)
			if err := os.Symlink(hdr.Linkname, destPath); err != nil {
				return err
			}
		}
	}
	return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `(cd core && go test -v ./pkg/store/oci -run TestExtractLayer_Basic)`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add core/pkg/store/oci/extractor.go core/pkg/store/oci/extractor_test.go
git commit -m "feat(ocistore): implement layer tar extraction with whiteout support"
```

---

### Task 5: Implement Mount with `fuse-overlayfs` and Release Logic

**Files:**
- Create: `core/pkg/store/oci/oras_mount.go`
- Modify: `core/pkg/store/oci/oras.go`
- Test: `core/pkg/store/oci/oras_mount_test.go`

**Interfaces:**
- Consumes: `fuse-overlayfs`, `fusermount3` / `fusermount`, `orasImage`
- Produces: `(s *orasStore) Mount(ctx context.Context, image Image, opts ...MountOption) (*Mount, error)`

- [ ] **Step 1: Write the failing test for Mount parameter building and resolution**

Create `core/pkg/store/oci/oras_mount_test.go`:
```go
/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ocistore

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	coreconfig "drassi.run/core/config"
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `(cd core && go test -v ./pkg/store/oci -run TestBuildFuseOverlayArgs)`
Expected: FAIL with undefined `buildFuseOverlayArgs`.

- [ ] **Step 3: Implement Mount and Release execution**

Create `core/pkg/store/oci/oras_mount.go`:
- Binary lookup for `fuse-overlayfs`: user config path or `exec.LookPath("fuse-overlayfs")`.
- Unpack check for each image layer: stream from `s.store.Fetch` if `layers/<diffID>` does not exist yet.
- Directory creation for `target` and `snapshots/<uuid>/{upper,work}`.
- Running command `exec.CommandContext(ctx, binary, args...)`.
- Release closure calling `unmountFUSE(target)`:
  - Tries `fusermount3 -u <target>`.
  - Fallback `fusermount -u <target>`.
  - Removes snapshot directories.
  - Wrapped in `xsync.InvokeOnce`.

- [ ] **Step 4: Run test to verify it passes**

Run: `(cd core && go test -v ./pkg/store/oci -run TestBuildFuseOverlayArgs)`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add core/pkg/store/oci/oras_mount.go core/pkg/store/oci/oras_mount_test.go
git commit -m "feat(ocistore): implement fuse-overlayfs mount and fusermount release"
```

---

### Task 6: Implement Read Workflow and Integration Testing

**Files:**
- Modify: `core/pkg/store/oci/oras.go`
- Test: `core/pkg/store/oci/oras_integration_test.go`

**Interfaces:**
- Consumes: `Mount`, `xfs.Read`, `osfs.New`
- Produces: `(s *orasStore) Read(ctx context.Context, image Image, opts ...ReadOption) (io.ReadCloser, error)`

- [ ] **Step 1: Write integration test for oras store workflow**

Create `core/pkg/store/oci/oras_integration_test.go`:
```go
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
	require.NoError(t, err)
	defer mnt.Release(ctx)

	// Read
	rc, err := mgr.Read(ctx, img, WithSubpath("hello.txt"))
	require.NoError(t, err)
	defer rc.Close()

	readData, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.Equal(t, content, readData)
}
```

- [ ] **Step 2: Run test to verify it fails before Read is completed**

Run: `(cd core && go test -v ./pkg/store/oci -run TestOrasStore_EndToEnd)`
Expected: FAIL

- [ ] **Step 3: Implement Read method in `core/pkg/store/oci/oras.go`**

Implement `Read` using temporary directory, `Mount`, `osfs.New(mnt.Target)`, and `xfs.Read(ctx, fsys, ro.Subpath)`.

- [ ] **Step 4: Run all tests to verify they pass**

Run: `(cd core && go test -v ./pkg/store/oci/...)`
Expected: PASS

- [ ] **Step 5: Run linting and hawkeye format**

Run: `golangci-lint run ./core/... && hawkeye format`

- [ ] **Step 6: Commit**

```bash
git add core/pkg/store/oci/oras.go core/pkg/store/oci/oras_integration_test.go
git commit -m "feat(ocistore): implement Read method and end-to-end integration tests"
```
