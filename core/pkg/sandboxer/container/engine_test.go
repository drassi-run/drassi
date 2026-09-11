/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package container

import (
	"context"
	"testing"

	"drassi.run/core/config"
	mock_container "drassi.run/core/mock/container"
	mock_store "drassi.run/core/mock/store/oci"
	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/sandboxer"
	ocistore "drassi.run/core/pkg/store/oci"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

func TestContainerEngineSuite(t *testing.T) {
	suite.Run(t, new(ContainerEngineTestSuite))
}

type ContainerEngineTestSuite struct {
	suite.Suite
	ctrl       *gomock.Controller
	store      *mock_store.MockManager
	mockClient *mock_container.MockEngine
}

func (s *ContainerEngineTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.store = mock_store.NewMockManager(s.ctrl)
	s.mockClient = mock_container.NewMockEngine(s.ctrl)
}

func (s *ContainerEngineTestSuite) mockContainerLifecycle(containerID string, validateSpec func(spec *types.ContainerSpec)) {
	s.mockClient.EXPECT().ContainerRun(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, spec *types.ContainerSpec, _ *container.RunOptions) (string, error) {
			if validateSpec != nil {
				validateSpec(spec)
			}
			return containerID, nil
		},
	)
	s.mockClient.EXPECT().CopyIn(gomock.Any(), containerID, gomock.Any()).Return(nil)
	s.mockClient.EXPECT().ContainerInspect(gomock.Any(), containerID).Return(&types.ContainerSpec{}, nil)
	s.mockClient.EXPECT().ContainerRemove(gomock.Any(), gomock.Any()).Return(nil)
}

func (s *ContainerEngineTestSuite) assertLaunch(eng sandboxer.Engine, expectedID string) sandboxer.Sandbox {
	resp, err := eng.Launch(s.T().Context(), &sandboxer.LaunchRequest{})
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().NotNil(resp.Sandbox)
	s.Require().NotNil(resp.JobContainer)
	s.Require().Equal(expectedID, resp.JobContainer.Id)
	return resp.Sandbox
}

func (s *ContainerEngineTestSuite) TestLaunch() {
	s.Run("with provisioner", func() {
		img := &ocistore.Image{}
		s.store.EXPECT().Image(gomock.Any(), "drassi/node:24").Return(img, nil).AnyTimes()
		s.store.EXPECT().Mount(gomock.Any(), img, gomock.Any()).Return("/var/lib/drassi/node_mount", "layer-node", nil).AnyTimes()
		s.store.EXPECT().Unmount(gomock.Any(), "layer-node").Return(nil).AnyTimes()

		runtimes := map[string]*config.Runtime{
			"node": {Image: "drassi/node:24"},
		}

		p, err := NewProvisioner(s.store, runtimes)
		s.Require().NoError(err)
		s.Require().NotNil(p)

		s.mockContainerLifecycle("c-123", func(spec *types.ContainerSpec) {
			s.Require().Len(spec.Mounts, 1)
			s.Require().Equal("/var/lib/drassi/node_mount", spec.Mounts[0].Source)
			s.Require().Equal("/opt/drassi/runtimes/node", spec.Mounts[0].Target)
		})

		eng := New(s.mockClient, "default:image", p)
		sb := s.assertLaunch(eng, "c-123")
		s.Require().NoError(sb.Terminate(s.T().Context()))
	})

	s.Run("without provisioner", func() {
		s.mockContainerLifecycle("c-456", func(spec *types.ContainerSpec) {
			s.Require().Empty(spec.Mounts)
		})

		eng := New(s.mockClient, "default:image", nil)
		sb := s.assertLaunch(eng, "c-456")
		s.Require().NoError(sb.Terminate(s.T().Context()))
	})
}

func (s *ContainerEngineTestSuite) TestNewProvisioner() {
	s.Run("with runtimes and store", func() {
		runtimes := map[string]*config.Runtime{
			"node": {Image: "drassi/node:24"},
		}
		p, err := NewProvisioner(s.store, runtimes)
		s.Require().NoError(err)
		s.Require().NotNil(p)
	})

	s.Run("with runtimes but missing store returns error", func() {
		runtimes := map[string]*config.Runtime{
			"node": {Image: "drassi/node:24"},
		}
		p, err := NewProvisioner(nil, runtimes)
		s.Require().Error(err)
		s.Require().Nil(p)
		s.Require().Contains(err.Error(), "oci store is required")
	})

	s.Run("without runtimes returns nil provisioner", func() {
		p, err := NewProvisioner(s.store, nil)
		s.Require().NoError(err)
		s.Require().Nil(p)
	})
}
