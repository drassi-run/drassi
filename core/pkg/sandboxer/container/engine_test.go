/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package container

import (
	"context"
	"errors"
	"testing"

	mock_container "drassi.run/core/mock/container"
	mock_sandboxer "drassi.run/core/mock/sandboxer"
	c "drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/sandboxer"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

func TestContainerizedEngineSuite(t *testing.T) {
	suite.Run(t, new(ContainerizedEngineTestSuite))
}

type ContainerizedEngineTestSuite struct {
	suite.Suite
	ctrl       *gomock.Controller
	mockEng    *mock_sandboxer.MockEngine
	mockSb     *mock_sandboxer.MockSandbox
	mockClient *mock_container.MockEngine
	forge      *records.Forge
	eng        sandboxer.Engine
}

func (s *ContainerizedEngineTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockEng = mock_sandboxer.NewMockEngine(s.ctrl)
	s.mockSb = mock_sandboxer.NewMockSandbox(s.ctrl)
	s.mockClient = mock_container.NewMockEngine(s.ctrl)
	s.forge = &records.Forge{
		Repository: "repo",
		Workflow:   "wf",
		Job:        "job",
		RunId:      "1",
		RunAttempt: "1",
	}
	s.eng = WithContainers(s.mockEng)
}

func (s *ContainerizedEngineTestSuite) SetupSubTest() {
	s.SetupTest()
}

func (s *ContainerizedEngineTestSuite) TestWrap() {
	s.Run("double wrap and unwrap", func() {
		wrapped := WithContainers(s.mockEng)
		s.Require().NotNil(wrapped)
		s.Require().Equal(wrapped, WithContainers(wrapped))

		unwrapper, ok := wrapped.(interface{ Unwrap() sandboxer.Engine })
		s.Require().True(ok)
		s.Require().Equal(s.mockEng, unwrapper.Unwrap())
	})
}

func (s *ContainerizedEngineTestSuite) TestLaunch() {
	s.Run("no containers", func() {
		req := &sandboxer.LaunchRequest{
			Forge: s.forge,
		}
		expectedResp := &sandboxer.LaunchResponse{Sandbox: s.mockSb}
		s.mockEng.EXPECT().Launch(gomock.Any(), req).Return(expectedResp, nil)

		resp, err := s.eng.Launch(s.T().Context(), req)
		s.Require().NoError(err)
		s.Require().Equal(expectedResp, resp)
	})

	s.Run("with job container", func() {
		req := &sandboxer.LaunchRequest{
			Forge:        s.forge,
			JobContainer: &workflows.Container{Image: "alpine"},
		}

		customMounts := []*types.Mount{
			{Type: "bind", Source: "/host/workspace", Target: "/job/workspace"},
			{Type: "bind", Source: "/host/temp", Target: "/job/temp"},
			{Type: "bind", Source: "/host/runtimes/node", Target: "/job/runtimes/node"},
		}
		mounter := mock_sandboxer.NewMockMounter(s.ctrl)
		mounter.EXPECT().OverlayMounts(gomock.Any(), true).Return(customMounts)

		s.mockEng.EXPECT().Launch(gomock.Any(), req).Return(&sandboxer.LaunchResponse{
			Sandbox: s.mockSb,
			Mounter: mounter,
			ContainerEngine: func(context.Context) (c.Engine, error) {
				return s.mockClient, nil
			},
		}, nil)

		s.mockSb.EXPECT().Layout().Return(sandboxer.StandardLayout("/opt/drassi/")).AnyTimes()

		s.mockClient.EXPECT().NetworkCreate(gomock.Any(), gomock.Any()).Return("net-123", nil)
		s.mockClient.EXPECT().Address().Return("unix:///var/run/docker.sock")
		s.mockClient.EXPECT().ImagePull(gomock.Any(), "alpine", gomock.Any()).Return(nil)

		var capturedSpec *types.ContainerSpec
		s.mockClient.EXPECT().ContainerRun(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, spec *types.ContainerSpec, _ *c.RunOptions) (string, error) {
				capturedSpec = spec
				return "c-123", nil
			},
		)

		s.mockClient.EXPECT().CopyIn(gomock.Any(), "c-123", gomock.Any()).Return(nil)
		s.mockClient.EXPECT().ContainerInspect(gomock.Any(), "c-123").Return(&types.ContainerSpec{
			Environment: map[string]string{"PATH": "/bin"},
		}, nil)

		resp, err := s.eng.Launch(s.T().Context(), req)
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Require().NotNil(resp.JobContainer)
		s.Require().Equal("c-123", resp.JobContainer.Id)

		s.Require().NotNil(capturedSpec)
		for _, m := range customMounts {
			s.Require().Contains(capturedSpec.Mounts, m)
		}

		s.mockClient.EXPECT().ContainerRemove(gomock.Any(), &c.RemoveOptions{Id: "c-123"}).Return(nil)
		s.mockClient.EXPECT().ContainerRemove(gomock.Any(), &c.RemoveOptions{Labels: s.forge.WellKnownLabels()}).Return(nil)
		s.mockClient.EXPECT().VolumeRemove(gomock.Any(), &c.RemoveOptions{Labels: s.forge.WellKnownLabels()}).Return(nil)
		s.mockClient.EXPECT().NetworkRemove(gomock.Any(), &c.RemoveOptions{Id: "net-123"}).Return(nil)
		s.mockSb.EXPECT().Terminate(gomock.Any()).Return(nil)

		s.Require().NoError(resp.Sandbox.Terminate(s.T().Context()))
	})

	s.Run("with service containers only", func() {
		req := &sandboxer.LaunchRequest{
			Forge: s.forge,
			ServiceContainers: map[string]*workflows.Container{
				"redis": {Image: "redis:7"},
			},
		}

		s.mockEng.EXPECT().Launch(gomock.Any(), req).Return(&sandboxer.LaunchResponse{
			Sandbox: s.mockSb,
			ContainerEngine: func(context.Context) (c.Engine, error) {
				return s.mockClient, nil
			},
		}, nil)

		s.mockClient.EXPECT().NetworkCreate(gomock.Any(), gomock.Any()).Return("net-svc", nil)
		s.mockClient.EXPECT().ImagePull(gomock.Any(), "redis:7", gomock.Any()).Return(nil)
		s.mockClient.EXPECT().ContainerRun(gomock.Any(), gomock.Any(), gomock.Any()).Return("c-redis", nil)
		s.mockClient.EXPECT().ContainerInspect(gomock.Any(), "c-redis").Return(&types.ContainerSpec{}, nil)

		resp, err := s.eng.Launch(s.T().Context(), req)
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Require().Nil(resp.JobContainer)
		s.Require().Contains(resp.ServiceContainers, "redis")
		s.Require().Equal("c-redis", resp.ServiceContainers["redis"].Id)
		s.Require().NotNil(resp.Sandbox)

		// Terminating wrapped underlay sandbox runs b.cleanups before underlay sandbox Terminate
		s.mockClient.EXPECT().ContainerRemove(gomock.Any(), &c.RemoveOptions{Labels: s.forge.WellKnownLabels()}).Return(nil)
		s.mockClient.EXPECT().VolumeRemove(gomock.Any(), &c.RemoveOptions{Labels: s.forge.WellKnownLabels()}).Return(nil)
		s.mockClient.EXPECT().NetworkRemove(gomock.Any(), &c.RemoveOptions{Id: "net-svc"}).Return(nil)
		s.mockSb.EXPECT().Terminate(gomock.Any()).Return(nil)

		s.Require().NoError(resp.Sandbox.Terminate(s.T().Context()))
	})

	s.Run("missing container engine", func() {
		s.mockSb.EXPECT().Terminate(gomock.Any()).Return(nil)

		req := &sandboxer.LaunchRequest{
			Forge:        s.forge,
			JobContainer: &workflows.Container{Image: "alpine"},
		}
		s.mockEng.EXPECT().Launch(gomock.Any(), req).Return(&sandboxer.LaunchResponse{
			Sandbox: s.mockSb,
		}, nil)

		resp, err := s.eng.Launch(s.T().Context(), req)
		s.Require().Nil(resp)
		s.Require().ErrorContains(err, "sandbox does not support container engine")
	})

	s.Run("container engine provider error", func() {
		s.mockSb.EXPECT().Terminate(gomock.Any()).Return(nil)

		req := &sandboxer.LaunchRequest{
			Forge:        s.forge,
			JobContainer: &workflows.Container{Image: "alpine"},
		}
		s.mockEng.EXPECT().Launch(gomock.Any(), req).Return(&sandboxer.LaunchResponse{
			Sandbox: s.mockSb,
			ContainerEngine: func(context.Context) (c.Engine, error) {
				return nil, errors.New("docker daemon down")
			},
		}, nil)

		resp, err := s.eng.Launch(s.T().Context(), req)
		s.Require().Nil(resp)
		s.Require().ErrorContains(err, "docker daemon down")
	})
}
