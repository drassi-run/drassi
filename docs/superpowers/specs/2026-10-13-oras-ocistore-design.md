# Design Specification: `ocistore.Manager` using ORAS-Go and FUSE-OverlayFS

## 1. Overview & Goal

Implement an alternative `ocistore.Manager` backend using `oras.land/oras-go/v2` and `fuse-overlayfs`.
This provides a rootless, self-contained OCI image storage and mounting system without requiring a running `containerd` daemon.

## 2. Architecture & File Layout

All store files reside in the base directory configured for OCI storage:
Default path: `filepath.Join(coreconfig.RootDir(), "ocistore")`

### Directory Layout

```
<RootDir>/ocistore/
  ├── contents/           # Standard OCI Image Layout managed by oras-go
  │   ├── oci-layout      # Standard OCI layout header
  │   ├── index.json      # OCI index mapping image references/tags to manifests
  │   └── blobs/          # Content-addressable blobs (manifests, configs, layers)
  │       └── sha256/
  ├── layers/<diffID>/    # Unpacked layer rootfs directories (cached by layer diffID)
  └── snapshots/<uuid>/   # Ephemeral upper and work directories for active mounts
      ├── upper/          # Overlay writable layer
      └── work/           # Overlay scratchpad work directory
```

## 3. Configuration Changes

### `core/config/config.go`
Add `OrasStoreConfig` and wire it into `OciStore`:

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

- When `StoreDir` is empty, it defaults to `filepath.Join(coreconfig.RootDir(), "ocistore")`.
- When `FuseOverlayfsBinary` is empty, it resolves via `exec.LookPath("fuse-overlayfs")`.
- When `OciStore.Backend == "oras"`, `ocistore.New(cfg)` instantiates `NewOrasStore(cfg.Oras)`.

## 4. Component Design

### 4.1 Interface Conformance

The new store conforms to `ocistore.Manager`:
- `Pull(ctx context.Context, imageRef string, opts ...PullOption) (Image, error)`
- `Mount(ctx context.Context, image Image, opts ...MountOption) (*Mount, error)`
- `Read(ctx context.Context, image Image, opts ...ReadOption) (io.ReadCloser, error)`
- `Close() error`

### 4.2 Pull Workflow (`oras.Copy`)
1. Respect `PullOption`:
   - `PullNever`: Resolve reference in `contents/`. If not present, return `ErrImageNotFound`.
   - `PullIfNotPresent`: Resolve reference in `contents/`. If found, return `Image` immediately.
   - `PullAlways`: Proceed to pull from remote.
2. Initialize remote repository target via `remote.NewRepository(parsedRef)`.
   - Support `WithAuth(user, pass)`.
   - Support `WithInsecureSkipTLSVerify(bool)`.
3. Copy from remote to local OCI target using `oras.Copy`.
4. Tag/index the image in local store under `imageRef`.
5. Return `Image` implementation (`orasImage`).

### 4.3 Unpack & Layer Cache
1. Parse image manifest and config (`ocispec.Image`).
2. Resolve ordered list of layers (`RootFS.DiffIDs` and layer descriptors).
3. For each layer:
   - Check if `layers/<diffID>` exists.
   - If missing, fetch layer blob reader from `contents/`.
   - Extract archive into temporary directory `layers/.tmp-<uuid>` while respecting directory/file modes, symlinks, and whiteout files (`.wh.*`).
   - Atomically rename `layers/.tmp-<uuid>` to `layers/<diffID>`.
   - Synchronize unpack operations with a mutex/lock to prevent concurrent conflicting extractions of the same layer.

### 4.4 Mount Workflow (`fuse-overlayfs`)
1. Verify unpacked layers.
2. Order `lowerdir` parameter from top-most layer to base layer (`diffID_N:diffID_N-1:...:diffID_0`).
3. If `Writable: true`:
   - Allocate `snapshotID = uuid.NewString()`.
   - Create directories `<RootDir>/ocistore/snapshots/<snapshotID>/upper` and `<RootDir>/ocistore/snapshots/<snapshotID>/work`.
   - Command: `<fuse-overlayfs> -o lowerdir=<lowers>,upperdir=<upper>,workdir=<work> <targetDir>`
4. If `Writable: false`:
   - Command: `<fuse-overlayfs> -o lowerdir=<lowers> <targetDir>`
5. Execute `fuse-overlayfs` command. Ensure mountpoint exists before executing.
6. Provide `Mount.Release(ctx)`:
   - Run `fusermount3 -u <targetDir>` (falling back to `fusermount -u <targetDir>`).
   - Remove `<RootDir>/ocistore/snapshots/<snapshotID>` if writable.
   - Wrap in `xsync.InvokeOnce` for thread safety and idempotency.

### 4.5 Read Workflow
1. Create temporary directory under `/tmp/drassi-oci-read-*`.
2. Invoke `Mount(ctx, image, WithTarget(tempDir))`.
3. Open subpath using `osfs.New(tempDir)` and stream data.
4. Wrap stream in `readCloser` which releases mount and deletes temporary directory on `Close()`.

## 5. Security & Non-Root Guarantees
- No `sudo` or elevated root privileges are required.
- `fuse-overlayfs` runs as an unprivileged user space FUSE daemon.
- `fusermount3 -u` / `fusermount -u` cleans up user FUSE mounts without root permissions.

## 6. Testing Strategy
- Unit tests with mock/stub executables for `fuse-overlayfs` command execution and option parsing.
- Layer unpacking tests verifying tarball decompression and whiteout file handling.
- Integration test for `Pull`, `Mount`, `Read`, and `Release` when `fuse-overlayfs` is present on the host system.
