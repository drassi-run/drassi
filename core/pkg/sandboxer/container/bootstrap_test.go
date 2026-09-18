/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package container

import (
	"errors"
	"testing"

	mock_container "drassi.run/core/mock/container"
	mock_sandboxer "drassi.run/core/mock/sandboxer"
	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/specdef"
	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/sandboxer"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

func TestBootstrapperSuite(t *testing.T) {
	suite.Run(t, new(BootstrapTestSuite))
}

type BootstrapTestSuite struct {
	suite.Suite
	ctrl       *gomock.Controller
	mockClient *mock_container.MockEngine
	mockSb     *mock_sandboxer.MockSandbox
	layout     sandboxer.Layout
	forge      *records.Forge
	b          *Bootstrapper
}

func (s *BootstrapTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockClient = mock_container.NewMockEngine(s.ctrl)
	s.mockSb = mock_sandboxer.NewMockSandbox(s.ctrl)
	s.layout = DefaultLayout
	s.forge = &records.Forge{
		Repository: "drassi/test",
		Workflow:   "test.yml",
		Job:        "build",
		RunId:      "100",
		RunAttempt: "1",
	}
	s.b = NewBootstrapper(s.mockClient, s.forge)
}

func (s *BootstrapTestSuite) SetupSubTest() {
	s.SetupTest()
}

func (s *BootstrapTestSuite) TestNetwork() {
	s.Run("non-ephemeral", func() {
		s.b.Network("pre-existing-net", false)

		netId, err := s.b.ensureNetwork(s.T().Context())
		s.Require().NoError(err)
		s.Require().Equal("pre-existing-net", netId)
		// cleanups should not contain network removal (only container and volume remove)
		s.Require().Len(s.b.cleanups, 2)
	})

	s.Run("ephemeral", func() {
		s.b.Network("ephemeral-net", true)

		netId, err := s.b.ensureNetwork(s.T().Context())
		s.Require().NoError(err)
		s.Require().Equal("ephemeral-net", netId)
		// cleanups should contain container, volume, and network remove
		s.Require().Len(s.b.cleanups, 3)
	})

	s.Run("memoized", func() {
		// NetworkCreate must only be called once even if ensureNetwork is called twice
		s.mockClient.EXPECT().NetworkCreate(gomock.Any(), gomock.Any()).Return("lazy-net-1", nil).Times(1)

		netId1, err := s.b.ensureNetwork(s.T().Context())
		s.Require().NoError(err)
		s.Require().Equal("lazy-net-1", netId1)

		netId2, err := s.b.ensureNetwork(s.T().Context())
		s.Require().NoError(err)
		s.Require().Equal("lazy-net-1", netId2)
		s.Require().Len(s.b.cleanups, 3)
	})
}

func (s *BootstrapTestSuite) TestRollback() {
	s.Run("success", func() {
		s.b.Network("net-to-rollback", true)

		s.mockClient.EXPECT().ContainerRemove(gomock.Any(), &container.RemoveOptions{Labels: s.forge.WellKnownLabels()}).Return(nil)
		s.mockClient.EXPECT().VolumeRemove(gomock.Any(), &container.RemoveOptions{Labels: s.forge.WellKnownLabels()}).Return(nil)
		s.mockClient.EXPECT().NetworkRemove(gomock.Any(), &container.RemoveOptions{Id: "net-to-rollback"}).Return(nil)

		err := s.b.Rollback(s.T().Context())
		s.Require().NoError(err)
	})

	s.Run("error handling", func() {
		s.b.Network("net-to-rollback", true)

		s.mockClient.EXPECT().ContainerRemove(gomock.Any(), gomock.Any()).Return(errors.New("container err"))
		s.mockClient.EXPECT().VolumeRemove(gomock.Any(), gomock.Any()).Return(nil)
		s.mockClient.EXPECT().NetworkRemove(gomock.Any(), gomock.Any()).Return(errors.New("network err"))

		err := s.b.Rollback(s.T().Context())
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "container err")
		s.Require().Contains(err.Error(), "network err")
	})
}

func (s *BootstrapTestSuite) TestRunJobContainer() {
	s.Run("with container def", func() {
		def := &workflows.Container{Image: "node:18"}

		s.mockClient.EXPECT().NetworkCreate(gomock.Any(), gomock.Any()).Return("net-job", nil)
		s.mockClient.EXPECT().Address().Return("unix:///var/run/docker.sock")
		s.mockClient.EXPECT().ImagePull(gomock.Any(), "node:18", gomock.Any()).Return(nil)
		s.mockClient.EXPECT().ContainerRun(gomock.Any(), gomock.Any(), gomock.Any()).Return("c-job", nil)

		mount := &types.Mount{Type: "bind", Source: "/workspace", Target: s.layout.Workspace()}
		info, err := s.b.RunJobContainer(s.T().Context(), def, specdef.AddMount(mount))
		s.Require().NoError(err)
		s.Require().NotNil(info)
		s.Require().Equal("c-job", info.Id)
		s.Require().Equal("net-job", info.Network)
	})

	s.Run("nil def", func() {
		info, err := s.b.RunJobContainer(s.T().Context(), nil)
		s.Require().NoError(err)
		s.Require().Nil(info)
	})
}

func (s *BootstrapTestSuite) TestRunServiceContainers() {
	s.Run("with service defs", func() {
		defs := map[string]*workflows.Container{
			"redis": {Image: "redis:7"},
		}

		s.mockClient.EXPECT().NetworkCreate(gomock.Any(), gomock.Any()).Return("net-svc", nil)
		s.mockClient.EXPECT().ImagePull(gomock.Any(), "redis:7", gomock.Any()).Return(nil)
		s.mockClient.EXPECT().ContainerRun(gomock.Any(), gomock.Any(), gomock.Any()).Return("c-redis", nil)
		s.mockClient.EXPECT().ContainerInspect(gomock.Any(), "c-redis").Return(&types.ContainerSpec{
			Publish: []*types.PortBinding{
				{HostPort: 6379, ContainerPort: 6379, Protocol: "tcp"},
			},
		}, nil)

		res, err := s.b.RunServiceContainers(s.T().Context(), defs)
		s.Require().NoError(err)
		s.Require().Contains(res, "redis")
		s.Require().Equal("c-redis", res["redis"].Id)
		s.Require().Equal("6379", res["redis"].Ports["6379"])
	})

	s.Run("empty defs", func() {
		res, err := s.b.RunServiceContainers(s.T().Context(), nil)
		s.Require().NoError(err)
		s.Require().Empty(res)
	})
}

func (s *BootstrapTestSuite) TestWrapSandbox() {
	s.Run("overlay only", func() {
		mockOverlay := mock_sandboxer.NewMockSandbox(s.ctrl)
		sb := s.b.WrapSandbox(nil, mockOverlay)
		s.Require().NotNil(sb)
	})

	s.Run("underlay only", func() {
		sb := s.b.WrapSandbox(s.mockSb, nil)
		s.Require().NotNil(sb)
	})

	s.Run("both layers", func() {
		mockOverlay := mock_sandboxer.NewMockSandbox(s.ctrl)
		sb := s.b.WrapSandbox(s.mockSb, mockOverlay)
		s.Require().NotNil(sb)
	})

	s.Run("both nil", func() {
		s.Require().PanicsWithValue("both sandbox layers are nil", func() {
			s.b.WrapSandbox(nil, nil)
		})
	})
}
