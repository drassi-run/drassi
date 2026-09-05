package ocistore

import (
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
