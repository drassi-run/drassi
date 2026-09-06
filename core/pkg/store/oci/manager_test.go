package ocistore

import (
	"errors"
	"testing"

	mock_storage "drassi.run/core/mock/podman/storage"
	"github.com/stretchr/testify/suite"
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

func (s *ManagerTestSuite) TestMount() {
	s.Run("read-only base mount success", func() {
		img := &storage.Image{
			ID:       "img-123",
			TopLayer: "layer-top-456",
		}
		s.store.EXPECT().Mount("layer-top-456", "").Return("/var/lib/oci/mounts/layer-top-456", nil)

		mountDir, id, err := s.mgr.Mount(s.T().Context(), img)
		s.Require().NoError(err)
		s.Require().Equal("/var/lib/oci/mounts/layer-top-456", mountDir)
		s.Require().Equal("layer-top-456", id)
	})

	s.Run("read-only mount layer mount error", func() {
		img := &storage.Image{
			ID:       "img-123",
			TopLayer: "layer-top-456",
		}
		s.store.EXPECT().Mount("layer-top-456", "").Return("", errors.New("mount permission denied"))

		mountDir, id, err := s.mgr.Mount(s.T().Context(), img)
		s.Require().Error(err)
		s.Require().Empty(mountDir)
		s.Require().Empty(id)
		s.Require().Contains(err.Error(), "mount image: mount permission denied")
	})

	s.Run("writable COW mount success", func() {
		img := &storage.Image{
			ID:       "img-123",
			TopLayer: "layer-base-456",
		}
		cowLayer := &storage.Layer{
			ID:     "layer-cow-789",
			Parent: "layer-base-456",
		}

		s.store.EXPECT().CreateLayer("", "layer-base-456", []string(nil), "", true, (*storage.LayerOptions)(nil)).Return(cowLayer, nil)
		s.store.EXPECT().Mount("layer-cow-789", "").Return("/var/lib/oci/mounts/layer-cow-789", nil)

		mountDir, id, err := s.mgr.Mount(s.T().Context(), img, WithWritable(true))
		s.Require().NoError(err)
		s.Require().Equal("/var/lib/oci/mounts/layer-cow-789", mountDir)
		s.Require().Equal("layer-cow-789", id)
	})

	s.Run("writable COW mount create layer error", func() {
		img := &storage.Image{
			ID:       "img-123",
			TopLayer: "layer-base-456",
		}
		s.store.EXPECT().CreateLayer("", "layer-base-456", []string(nil), "", true, (*storage.LayerOptions)(nil)).Return(nil, errors.New("disk full"))

		mountDir, id, err := s.mgr.Mount(s.T().Context(), img, WithWritable(true))
		s.Require().Error(err)
		s.Require().Empty(mountDir)
		s.Require().Empty(id)
		s.Require().Contains(err.Error(), "create COW layer: disk full")
	})

	s.Run("writable COW mount mount error cleans up layer", func() {
		img := &storage.Image{
			ID:       "img-123",
			TopLayer: "layer-base-456",
		}
		cowLayer := &storage.Layer{
			ID:     "layer-cow-789",
			Parent: "layer-base-456",
		}

		s.store.EXPECT().CreateLayer("", "layer-base-456", []string(nil), "", true, (*storage.LayerOptions)(nil)).Return(cowLayer, nil)
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

func (s *ManagerTestSuite) TestClose() {
	s.store.EXPECT().Shutdown(false).Return([]string(nil), nil)

	err := s.mgr.Close()
	s.Require().NoError(err)
}
