/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ocistore

import (
	"archive/tar"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	mock_storage "drassi.run/core/mock/podman/storage"
	mock_io "drassi.run/core/mock/sdk/io"
	"github.com/stretchr/testify/suite"
	"go.podman.io/image/v5/signature"
	"go.podman.io/image/v5/types"
	"go.podman.io/storage"
	"go.uber.org/mock/gomock"
)

func TestManagerSuite(t *testing.T) {
	suite.Run(t, new(ManagerTestSuite))
}

type ManagerTestSuite struct {
	suite.Suite
	ctrl  *gomock.Controller
	store *mock_storage.MockStore
	mgr   *manager
}

func (s *ManagerTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.store = mock_storage.NewMockStore(s.ctrl)
	s.store.EXPECT().GraphOptions().Return(nil).AnyTimes()
	s.store.EXPECT().GraphDriverName().Return("overlay").AnyTimes()
	s.store.EXPECT().GraphRoot().Return("/var/lib/containers/storage").AnyTimes()
	s.store.EXPECT().RunRoot().Return("/run/containers/storage").AnyTimes()

	s.mgr = New(s.store).(*manager)
}

func (s *ManagerTestSuite) newPolicyContext() *signature.PolicyContext {
	s.T().Helper()
	policy := &signature.Policy{
		Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()},
	}
	pCtx, err := signature.NewPolicyContext(policy)
	s.Require().NoError(err)
	s.T().Cleanup(func() { _ = pCtx.Destroy() })
	return pCtx
}

func (s *ManagerTestSuite) capturePullOptions(targetImg *storage.Image, opts ...PullOption) *pullOptions {
	s.T().Helper()
	var capturedOpts []PullOption
	s.mgr.pull = func(ctx context.Context, destRef, srcRef types.ImageReference, opts ...PullOption) error {
		capturedOpts = opts
		return nil
	}

	s.store.EXPECT().Image(gomock.Any()).Return(targetImg, nil)

	_, err := s.mgr.Pull(s.T().Context(), "alpine:3.20", opts...)
	s.Require().NoError(err)
	po := new(pullOptions)
	for _, opt := range capturedOpts {
		opt(po)
	}
	return po
}

func (s *ManagerTestSuite) mockCreateLayer(parent, layerID string, err error) *storage.Layer {
	s.T().Helper()
	var layer *storage.Layer
	if err == nil {
		layer = &storage.Layer{ID: layerID, Parent: parent}
	}
	s.store.EXPECT().CreateLayer("", parent, []string(nil), "", true, (*storage.LayerOptions)(nil)).Return(layer, err)
	return layer
}

func (s *ManagerTestSuite) mockImageMount(layerID, mountDir string) *storage.Image {
	s.T().Helper()
	s.store.EXPECT().Mount(layerID, "").Return(mountDir, nil)
	s.store.EXPECT().Unmount(layerID, false).Return(false, nil)
	return &storage.Image{ID: "img-" + layerID, TopLayer: layerID}
}

func (s *ManagerTestSuite) createTempDir(files, symlinks map[string]string) string {
	s.T().Helper()
	tempDir := s.T().TempDir()
	for relPath, content := range files {
		fullPath := filepath.Join(tempDir, relPath)
		s.Require().NoError(os.MkdirAll(filepath.Dir(fullPath), 0o755))
		s.Require().NoError(os.WriteFile(fullPath, []byte(content), 0o644))
	}
	for linkName, target := range symlinks {
		fullPath := filepath.Join(tempDir, linkName)
		s.Require().NoError(os.MkdirAll(filepath.Dir(fullPath), 0o755))
		s.Require().NoError(os.Symlink(target, fullPath))
	}
	return tempDir
}

func (s *ManagerTestSuite) assertTar(r io.Reader, expectedEntries, expectedSymlinks map[string]string) {
	s.T().Helper()
	entries := make(map[string]string)
	symlinks := make(map[string]string)
	tr := tar.NewReader(r)

	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		s.Require().NoError(err)

		if hdr.Typeflag == tar.TypeSymlink {
			symlinks[hdr.Name] = hdr.Linkname
		} else if hdr.Typeflag == tar.TypeReg {
			content, err := io.ReadAll(tr)
			s.Require().NoError(err)
			entries[hdr.Name] = string(content)
		}
	}

	if expectedEntries != nil {
		s.Require().Equal(expectedEntries, entries)
	} else {
		s.Require().Empty(entries)
	}

	if expectedSymlinks != nil {
		s.Require().Equal(expectedSymlinks, symlinks)
	} else {
		s.Require().Empty(symlinks)
	}
}

func (s *ManagerTestSuite) TestImage() {
	s.Run("found", func() {
		expectedImg := &storage.Image{
			ID:       "img-123",
			TopLayer: "layer-abc",
		}
		s.store.EXPECT().Image("alpine:latest").Return(expectedImg, nil)

		img, err := s.mgr.Image(s.T().Context(), "alpine:latest")
		s.Require().NoError(err)
		s.Require().Equal(expectedImg, img)
	})

	s.Run("not found", func() {
		s.store.EXPECT().Image("nonexistent:tag").Return(nil, storage.ErrImageUnknown)

		img, err := s.mgr.Image(s.T().Context(), "nonexistent:tag")
		s.Require().NoError(err)
		s.Require().Nil(img)
	})
}

func (s *ManagerTestSuite) TestPull() {
	s.Run("success singleflight concurrent", func() {
		expectedImg := &storage.Image{
			ID:       "img-pulled",
			TopLayer: "layer-pulled",
		}

		var copyCount atomic.Int32
		s.mgr.pull = func(ctx context.Context, destRef, srcRef types.ImageReference, opts ...PullOption) error {
			copyCount.Add(1)
			time.Sleep(50 * time.Millisecond) // simulate copy latency
			return nil
		}

		s.store.EXPECT().Image("alpine:latest").Return(expectedImg, nil).Times(1)

		const concurrency = 5
		var wg sync.WaitGroup

		wg.Add(concurrency)
		for i := 0; i < concurrency; i++ {
			go func() {
				defer wg.Done()
				img, err := s.mgr.Pull(s.T().Context(), "alpine:latest")
				s.Require().NoError(err)
				s.Require().Equal(expectedImg, img)
			}()
		}

		wg.Wait()
		s.Require().Equal(int32(1), copyCount.Load())
	})

	s.Run("options and context propagation", func() {
		expectedImg := &storage.Image{
			ID:       "img-custom",
			TopLayer: "layer-custom",
		}

		var (
			capturedSrcRef  types.ImageReference
			capturedDestRef types.ImageReference
			capturedOpts    []PullOption
		)

		s.mgr.pull = func(ctx context.Context, destRef, srcRef types.ImageReference, opts ...PullOption) error {
			capturedSrcRef = srcRef
			capturedDestRef = destRef
			capturedOpts = opts
			return nil
		}

		s.store.EXPECT().Image("docker.io/library/ubuntu:22.04").Return(expectedImg, nil)

		pCtx := s.newPolicyContext()
		mockWriter := mock_io.NewMockWriter(s.ctrl)

		img, err := s.mgr.Pull(
			s.T().Context(),
			"docker.io/library/ubuntu:22.04",
			WithAuth("testuser", "testpass"),
			WithAuthFile("/tmp/auth.json"),
			WithReportWriter(mockWriter),
			WithInsecureSkipTLSVerify(true),
			WithPlatform("linux", "arm64"),
			WithPolicyContext(pCtx),
		)
		s.Require().NoError(err)
		s.Require().Equal(expectedImg, img)

		s.Require().NotNil(capturedSrcRef)
		s.Require().NotNil(capturedDestRef)
		s.Require().NotEmpty(capturedOpts)

		po := new(pullOptions)
		for _, opt := range capturedOpts {
			opt(po)
		}
		s.Require().Equal(mockWriter, po.ReportWriter)
		s.Require().NotNil(po.SystemContext)
		s.Require().Equal("testuser", po.SystemContext.DockerAuthConfig.Username)
		s.Require().Equal("testpass", po.SystemContext.DockerAuthConfig.Password)
		s.Require().Equal("/tmp/auth.json", po.SystemContext.AuthFilePath)
		s.Require().Equal(types.OptionalBoolTrue, po.SystemContext.DockerInsecureSkipTLSVerify)
		s.Require().Equal("linux", po.SystemContext.OSChoice)
		s.Require().Equal("arm64", po.SystemContext.ArchitectureChoice)
		s.Require().Equal(pCtx, po.PolicyContext)

		resPC, ephemeral, err := po.PolicyCtx()
		s.Require().NoError(err)
		s.Require().False(ephemeral)
		s.Require().Equal(pCtx, resPC)
	})

	s.Run("individual arch and os options", func() {
		po := s.capturePullOptions(&storage.Image{ID: "img-alpine"}, WithArchitecture("riscv64"), WithOS("freebsd"))
		s.Require().Equal("riscv64", po.SystemContext.ArchitectureChoice)
		s.Require().Equal("freebsd", po.SystemContext.OSChoice)
	})

	s.Run("custom system context option", func() {
		customSys := &types.SystemContext{
			DockerRegistryUserAgent: "custom-agent",
		}
		po := s.capturePullOptions(&storage.Image{ID: "img-alpine"}, WithSystemContext(customSys))
		s.Require().Equal(customSys, po.SystemContext)
	})

	s.Run("invalid source reference", func() {
		img, err := s.mgr.Pull(s.T().Context(), ":::invalid reference:::")
		s.Require().Error(err)
		s.Require().Nil(img)
		s.Require().Contains(err.Error(), "parse source reference")
	})

	s.Run("copy error propagation", func() {
		s.mgr.pull = func(ctx context.Context, destRef, srcRef types.ImageReference, opts ...PullOption) error {
			return errors.New("network timeout")
		}

		img, err := s.mgr.Pull(s.T().Context(), "alpine:latest")
		s.Require().Error(err)
		s.Require().Nil(img)
		s.Require().Contains(err.Error(), "copy image \"alpine:latest\": network timeout")
	})

	s.Run("lookup after pull error", func() {
		s.mgr.pull = func(ctx context.Context, destRef, srcRef types.ImageReference, opts ...PullOption) error {
			return nil
		}

		s.store.EXPECT().Image("alpine:latest").Return(nil, errors.New("store corrupted"))

		img, err := s.mgr.Pull(s.T().Context(), "alpine:latest")
		s.Require().Error(err)
		s.Require().Nil(img)
		s.Require().Contains(err.Error(), "lookup pulled image \"alpine:latest\": store corrupted")
	})
}

func (s *ManagerTestSuite) TestPullOptions_PolicyCtx() {
	s.Run("default fallback policy when policy context is nil", func() {
		po := new(pullOptions)
		pc, ephemeral, err := po.PolicyCtx()
		s.Require().NoError(err)
		s.Require().True(ephemeral)
		s.Require().NotNil(pc)
		_ = pc.Destroy()
	})

	s.Run("explicit policy context returns non-ephemeral", func() {
		expectedPC := s.newPolicyContext()

		po := &pullOptions{PolicyContext: expectedPC}
		pc, ephemeral, err := po.PolicyCtx()
		s.Require().NoError(err)
		s.Require().False(ephemeral)
		s.Require().Equal(expectedPC, pc)
	})
}

func (s *ManagerTestSuite) TestMount() {
	img := &storage.Image{ID: "img-123", TopLayer: "layer-top-456"}

	s.Run("read-only base mount success", func() {
		s.store.EXPECT().Mount("layer-top-456", "").Return("/var/lib/oci/mounts/layer-top-456", nil)

		mountDir, id, err := s.mgr.Mount(s.T().Context(), img)
		s.Require().NoError(err)
		s.Require().Equal("/var/lib/oci/mounts/layer-top-456", mountDir)
		s.Require().Equal("layer-top-456", id)
	})

	s.Run("read-only mount layer mount error", func() {
		s.store.EXPECT().Mount("layer-top-456", "").Return("", errors.New("mount permission denied"))

		mountDir, id, err := s.mgr.Mount(s.T().Context(), img)
		s.Require().Error(err)
		s.Require().Empty(mountDir)
		s.Require().Empty(id)
		s.Require().Contains(err.Error(), "mount image: mount permission denied")
	})

	s.Run("writable COW mount success", func() {
		s.mockCreateLayer("layer-top-456", "layer-cow-789", nil)
		s.store.EXPECT().Mount("layer-cow-789", "").Return("/var/lib/oci/mounts/layer-cow-789", nil)

		mountDir, id, err := s.mgr.Mount(s.T().Context(), img, WithWritable(true))
		s.Require().NoError(err)
		s.Require().Equal("/var/lib/oci/mounts/layer-cow-789", mountDir)
		s.Require().Equal("layer-cow-789", id)
	})

	s.Run("writable COW mount create layer error", func() {
		s.mockCreateLayer("layer-top-456", "layer-cow-789", errors.New("disk full"))

		mountDir, id, err := s.mgr.Mount(s.T().Context(), img, WithWritable(true))
		s.Require().Error(err)
		s.Require().Empty(mountDir)
		s.Require().Empty(id)
		s.Require().Contains(err.Error(), "create COW layer: disk full")
	})

	s.Run("writable COW mount mount error cleans up layer", func() {
		s.mockCreateLayer("layer-top-456", "layer-cow-789", nil)
		s.store.EXPECT().Mount("layer-cow-789", "").Return("", errors.New("failed to mount"))
		s.store.EXPECT().DeleteLayer("layer-cow-789").Return(nil)

		mountDir, id, err := s.mgr.Mount(s.T().Context(), img, WithWritable(true))
		s.Require().Error(err)
		s.Require().Empty(mountDir)
		s.Require().Empty(id)
		s.Require().Contains(err.Error(), "mount image: failed to mount")
	})
}

func (s *ManagerTestSuite) TestUnmount() {
	s.Run("success", func() {
		s.store.EXPECT().Unmount("layer-cow-123", true).Return(false, nil)
		s.store.EXPECT().DeleteLayer("layer-cow-123").Return(nil)

		err := s.mgr.Unmount(s.T().Context(), "layer-cow-123")
		s.Require().NoError(err)
	})

	s.Run("unmount error propagates", func() {
		s.store.EXPECT().Unmount("layer-cow-123", true).Return(true, errors.New("device busy"))
		s.store.EXPECT().DeleteLayer("layer-cow-123").Return(nil)

		err := s.mgr.Unmount(s.T().Context(), "layer-cow-123")
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "device busy")
	})
}

func (s *ManagerTestSuite) TestRead() {
	s.Run("full image success", func() {
		tempDir := s.createTempDir(
			map[string]string{"file1.txt": "content1", "sub/file2.txt": "content2"},
			map[string]string{"link1": "file1.txt"},
		)
		img := s.mockImageMount("layer-read-1", tempDir)

		rc, err := s.mgr.Read(s.T().Context(), img)
		s.Require().NoError(err)
		defer rc.Close()

		s.assertTar(rc, map[string]string{
			"file1.txt":     "content1",
			"sub/file2.txt": "content2",
		}, map[string]string{
			"link1": "file1.txt",
		})
	})

	s.Run("subpath directory success", func() {
		tempDir := s.createTempDir(map[string]string{
			"app/config/settings.json": `{"port":8080}`,
			"app/config/app.conf":      "key=val",
		}, nil)
		img := s.mockImageMount("layer-read-2", tempDir)

		rc, err := s.mgr.Read(s.T().Context(), img, WithSubpath("app/config"))
		s.Require().NoError(err)
		defer rc.Close()

		s.assertTar(rc, map[string]string{
			"settings.json": `{"port":8080}`,
			"app.conf":      "key=val",
		}, nil)
	})

	s.Run("subpath single file success", func() {
		tempDir := s.createTempDir(map[string]string{"app/config.yaml": "env: test"}, nil)
		img := s.mockImageMount("layer-read-3", tempDir)

		rc, err := s.mgr.Read(s.T().Context(), img, WithSubpath("/app/config.yaml"))
		s.Require().NoError(err)
		defer rc.Close()

		s.assertTar(rc, map[string]string{"config.yaml": "env: test"}, nil)
	})

	s.Run("subpath not found error unmounts layer", func() {
		img := s.mockImageMount("layer-read-4", s.T().TempDir())

		rc, err := s.mgr.Read(s.T().Context(), img, WithSubpath("nonexistent"))
		s.Require().Error(err)
		s.Require().Nil(rc)
		s.Require().Contains(err.Error(), "resolve subpath \"nonexistent\"")
	})

	s.Run("subpath escapes root error unmounts layer", func() {
		img := s.mockImageMount("layer-read-5", s.T().TempDir())

		rc, err := s.mgr.Read(s.T().Context(), img, WithSubpath("../../etc/passwd"))
		s.Require().Error(err)
		s.Require().Nil(rc)
		s.Require().Contains(err.Error(), "resolve subpath \"../../etc/passwd\"")
	})

	s.Run("mount error", func() {
		img := &storage.Image{ID: "img-read-6", TopLayer: "layer-read-6"}
		s.store.EXPECT().Mount("layer-read-6", "").Return("", errors.New("mount failed"))

		rc, err := s.mgr.Read(s.T().Context(), img)
		s.Require().Error(err)
		s.Require().Nil(rc)
		s.Require().Contains(err.Error(), "mount image: mount failed")
	})

	s.Run("context cancellation during read", func() {
		tempDir := s.createTempDir(map[string]string{"file.txt": "data"}, nil)
		img := s.mockImageMount("layer-read-7", tempDir)

		ctx, cancel := context.WithCancel(s.T().Context())
		cancel() // cancel immediately

		rc, err := s.mgr.Read(ctx, img)
		s.Require().NoError(err)
		defer rc.Close()

		_, readErr := io.ReadAll(rc)
		s.Require().Error(readErr)
		s.Require().True(errors.Is(readErr, context.Canceled) || errors.Is(readErr, io.ErrClosedPipe))
	})
}

func (s *ManagerTestSuite) TestClose() {
	s.store.EXPECT().Shutdown(false).Return([]string(nil), nil)

	err := s.mgr.Close()
	s.Require().NoError(err)
}
