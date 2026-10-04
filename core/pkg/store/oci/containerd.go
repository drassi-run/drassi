/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ocistore

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"drassi.run/core/config"
	"drassi.run/core/util/error"
	"drassi.run/core/util/fs"
	"drassi.run/core/util/sync"
	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/leases"
	"github.com/containerd/containerd/v2/core/mount"
	"github.com/containerd/containerd/v2/core/remotes/docker"
	"github.com/containerd/errdefs"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/google/uuid"
	"github.com/opencontainers/image-spec/identity"
)

type ContainerdStoreConfig = config.ContainerdStoreConfig

// NewContainerdStore creates a new containerd-backed Manager.
func NewContainerdStore(cfg *config.ContainerdStoreConfig) (Manager, error) {
	if cfg == nil {
		cfg = config.DefaultContainerdStoreConfig()
	}
	var opts []containerd.Opt
	if cfg.Namespace != "" {
		opts = append(opts, containerd.WithDefaultNamespace(cfg.Namespace))
	}

	client, err := containerd.New(cfg.Address, opts...)
	if err != nil {
		return nil, fmt.Errorf("connect to containerd at %s: %w", cfg.Address, err)
	}

	m := &containerdStore{
		client:           client,
		fsSnapshotter:    cfg.FsSnapshotter,
		blockSnapshotter: cfg.BlockSnapshotter,
	}
	return m, nil
}

type containerdStore struct {
	client *containerd.Client

	fsSnapshotter    string // e.g overlayfs
	blockSnapshotter string // e.g blockfile | devmapper
}

var _ Manager = (*containerdStore)(nil)

func (s *containerdStore) Pull(ctx context.Context, imageRef string, opts ...PullOption) (Image, error) {
	po := &pullOptions{
		Policy: PullIfNotPresent,
	}
	for _, opt := range opts {
		opt(po)
	}

	if po.Policy == PullIfNotPresent || po.Policy == PullNever {
		img, err := s.client.GetImage(ctx, imageRef)
		if err == nil {
			return img, nil
		}
		if !errdefs.IsNotFound(err) {
			return nil, fmt.Errorf("get image %q: %w", imageRef, err)
		}
		if po.Policy == PullNever {
			return nil, ErrImageNotFound
		}
	}

	if leaseCtx, done, err := s.client.WithLease(ctx, leases.WithExpiration(2*time.Hour)); err != nil {
		return nil, fmt.Errorf("create lease for pull %q: %w", imageRef, err)
	} else {
		defer done(context.WithoutCancel(ctx)) //nolint:errcheck
		ctx = leaseCtx
	}

	var remoteOpts []containerd.RemoteOpt
	var regOpts []docker.RegistryOpt
	if po.InsecureSkipTLSVerify {
		transport := docker.DefaultHTTPTransport(&tls.Config{
			InsecureSkipVerify: true, //nolint:gosec
		})

		regOpts = append(regOpts, docker.WithClient(&http.Client{Transport: transport}))
	}
	if po.AuthUsername != "" || po.AuthPassword != "" {
		authorizer := docker.NewDockerAuthorizer(
			docker.WithAuthCreds(func(string) (string, string, error) {
				return po.AuthUsername, po.AuthPassword, nil
			}),
		)
		regOpts = append(regOpts, docker.WithAuthorizer(authorizer))
	}
	if len(regOpts) > 0 {
		resolver := docker.NewResolver(docker.ResolverOptions{
			Hosts: docker.ConfigureDefaultRegistries(regOpts...),
		})
		remoteOpts = append(remoteOpts, containerd.WithResolver(resolver))
	}

	img, err := s.client.Pull(ctx, imageRef, remoteOpts...)
	if err != nil {
		return nil, fmt.Errorf("pull image %q: %w", imageRef, err)
	}
	return img, nil
}

func (s *containerdStore) Mount(ctx context.Context, image Image, opts ...MountOption) (*Mount, error) {
	mo := &mountOptions{
		Writable: false,
	}
	for _, opt := range opts {
		opt(mo)
	}

	if (mo.Target != "") == mo.Block {
		return nil, errors.New("mount requires either block mode or target directory")
	}

	// Unpack
	var snapshotterName string
	if mo.Block {
		if s.blockSnapshotter == "" {
			return nil, errors.New("block snapshotter not configured")
		}
		snapshotterName = s.blockSnapshotter
	} else {
		snapshotterName = s.fsSnapshotter
	}
	if err := s.ensureUnpack(ctx, image, snapshotterName); err != nil {
		return nil, err
	}

	// Parent layer
	diffIDs, err := image.RootFS(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve image rootfs diff ids: %w", err)
	}
	parent := identity.ChainID(diffIDs).String()

	var rb xerror.Rollbacker
	defer rb.Run(ctx)

	// Lease
	leaser := s.client.LeasesService()
	lease, err := leaser.Create(ctx, leases.WithRandomID(), leases.WithExpiration(24*time.Hour))
	if err != nil {
		return nil, fmt.Errorf("create lease for mount: %w", err)
	}
	rb.Add(func(ctx context.Context) error {
		return leaser.Delete(ctx, lease)
	})
	ctx = leases.WithLease(ctx, lease.ID)

	// Snapshot
	snapshotter := s.client.SnapshotService(snapshotterName)
	snapshotKey := uuid.NewString()
	var mounts []mount.Mount
	if mo.Writable {
		mounts, err = snapshotter.Prepare(ctx, snapshotKey, parent)
	} else {
		mounts, err = snapshotter.View(ctx, snapshotKey, parent)
	}
	if err != nil {
		return nil, fmt.Errorf("prepare snapshot %s: %w", snapshotKey, err)
	}
	rb.Add(func(ctx context.Context) error {
		return snapshotter.Remove(ctx, snapshotKey)
	})

	// Block file or device
	if mo.Block {
		if len(mounts) != 1 {
			return nil, fmt.Errorf("expected single block mount, got %d", len(mounts))
		}
		rb.Dismiss() // success
		mnt := &Mount{
			Target:  mounts[0].Source,
			Release: xsync.InvokeOnce(rb.Do),
		}
		return mnt, nil
	}

	// Mount filesystem
	if err = mount.All(mounts, mo.Target); err != nil {
		return nil, fmt.Errorf("mount layers to %s: %w", mo.Target, err)
	}
	rb.Add(func(ctx context.Context) error {
		return mount.UnmountAll(mo.Target, 0)
	})

	rb.Dismiss() // success
	mnt := &Mount{
		Target:  mo.Target,
		Release: xsync.InvokeOnce(rb.Do),
	}
	return mnt, nil
}

func (s *containerdStore) ensureUnpack(ctx context.Context, image Image, snapshotter string) (err error) {
	img, ok := image.(containerd.Image)
	if !ok {
		img, err = s.client.GetImage(ctx, image.Name())
		if err != nil {
			return fmt.Errorf("get containerd image %q: %w", image.Name(), err)
		}
	}

	unpacked, err := img.IsUnpacked(ctx, snapshotter)
	if err != nil {
		return fmt.Errorf("check image unpacked %q: %w", image.Name(), err)
	}
	if !unpacked {
		if err = img.Unpack(ctx, snapshotter); err != nil {
			return fmt.Errorf("unpack image %q with snapshotter %q: %w", image.Name(), snapshotter, err)
		}
	}

	return nil
}

func (s *containerdStore) Read(ctx context.Context, image Image, opts ...ReadOption) (io.ReadCloser, error) {
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

func (s *containerdStore) Close() error {
	return s.client.Close()
}
