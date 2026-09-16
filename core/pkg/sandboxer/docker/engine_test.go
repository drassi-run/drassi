/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package docker

import (
	"context"
	"errors"
	"testing"

	"drassi.run/core/config"
	mock_container "drassi.run/core/mock/container"
	mock_store "drassi.run/core/mock/store/oci"
	c "drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/sandboxer/container"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	require.NotNil(t, cfg)
	require.NotNil(t, cfg.Template)
	require.Equal(t, container.DefaultImage, cfg.Template.Image)
	require.Equal(t, "unix:///var/run/docker.sock", cfg.Endpoint)
}

func TestFactoryRegistration(t *testing.T) {
	sbConfig := &config.Sandboxer{
		Provider: config.ProviderDocker,
	}
	f, err := sandboxer.NewFactory(sbConfig)
	require.NoError(t, err)
	require.NotNil(t, f)
}

func TestFactoryWithTemplateConfig(t *testing.T) {
	rawToml := `
endpoint = "unix:///var/run/docker.sock"
[template]
image = "ghcr.io/drassi-run/ubuntu:26.04"
network_mode = "host"
privileged = true
user = "1000:1000"
volumes = [
    "/var/run/docker.sock:/var/run/docker.sock",
    "/data:/data:ro"
]
ports = [
    "8080:80"
]
cap_add = ["SYS_ADMIN"]
devices = ["/dev/kvm"]
cpus = "2"
memory = 2147483648
`
	sbConfig := &config.Sandboxer{
		Provider: config.ProviderDocker,
		Config:   []byte(rawToml),
	}
	f, err := sandboxer.NewFactory(sbConfig)
	require.NoError(t, err)
	require.NotNil(t, f)

	fact, ok := f.(*factory)
	require.True(t, ok)
	require.Equal(t, "unix:///var/run/docker.sock", fact.cfg.Endpoint)
	require.NotNil(t, fact.cfg.Template)
	tmpl := fact.cfg.Template
	require.Equal(t, "ghcr.io/drassi-run/ubuntu:26.04", tmpl.Image)
	require.Equal(t, "host", tmpl.NetworkMode)
	require.True(t, tmpl.Privileged)
	require.Equal(t, "1000:1000", tmpl.User)
	require.Len(t, tmpl.Mounts, 2)
	require.Equal(t, "/var/run/docker.sock", tmpl.Mounts[0].Source)
	require.Equal(t, "/var/run/docker.sock", tmpl.Mounts[0].Target)
	require.False(t, tmpl.Mounts[0].ReadOnly)
	require.Equal(t, "/data", tmpl.Mounts[1].Source)
	require.Equal(t, "/data", tmpl.Mounts[1].Target)
	require.True(t, tmpl.Mounts[1].ReadOnly)
	require.Len(t, tmpl.Publish, 1)
	require.Equal(t, uint16(8080), tmpl.Publish[0].HostPort)
	require.Equal(t, uint16(80), tmpl.Publish[0].ContainerPort)
	require.Equal(t, []string{"SYS_ADMIN"}, tmpl.CapAdd)
	require.Equal(t, []string{"/dev/kvm"}, tmpl.Devices)
	require.Equal(t, "2", tmpl.CPUS)
	require.Equal(t, types.UnitBytes(2147483648), tmpl.Memory)
}

func TestFactoryWithShortFormTemplateConfig(t *testing.T) {
	rawToml := `
endpoint = "unix:///var/run/docker.sock"
[template]
image = "ghcr.io/drassi-run/ubuntu:26.04"
environment = [
    "APP_ENV=production",
    "DEBUG=false"
]
labels = [
    "org.drassi.env=prod"
]
annotations = [
    "note=short-form"
]
sysctls = [
    "net.ipv4.ip_forward=1"
]
expose = [
    "80/tcp",
    53
]
memory = "4g"
shm_size = "256m"
`
	sbConfig := &config.Sandboxer{
		Provider: config.ProviderDocker,
		Config:   []byte(rawToml),
	}
	f, err := sandboxer.NewFactory(sbConfig)
	require.NoError(t, err)
	require.NotNil(t, f)

	fact, ok := f.(*factory)
	require.True(t, ok)
	require.NotNil(t, fact.cfg.Template)
	tmpl := fact.cfg.Template
	require.Equal(t, "production", tmpl.Environment["APP_ENV"])
	require.Equal(t, "false", tmpl.Environment["DEBUG"])
	require.Equal(t, "prod", tmpl.Labels["org.drassi.env"])
	require.Equal(t, "short-form", tmpl.Annotations["note"])
	require.Equal(t, "1", tmpl.Sysctls["net.ipv4.ip_forward"])
	require.Len(t, tmpl.Exposes, 2)
	require.Equal(t, uint16(80), tmpl.Exposes[0].Number)
	require.Equal(t, "tcp", tmpl.Exposes[0].Protocol)
	require.Equal(t, uint16(53), tmpl.Exposes[1].Number)
	require.Equal(t, "tcp", tmpl.Exposes[1].Protocol)
	require.Equal(t, types.UnitBytes(4*1024*1024*1024), tmpl.Memory)
	require.Equal(t, types.UnitBytes(256*1024*1024), tmpl.ShmSize)
}

func TestFactoryCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mock_store.NewMockManager(ctrl)

	t.Run("without runtimes creates engine successfully", func(t *testing.T) {
		f := NewFactory(DefaultConfig())
		eng, err := f.Create()
		require.NoError(t, err)
		require.NotNil(t, eng)
		_ = eng.Close()
	})

	t.Run("with runtimes and store creates engine successfully", func(t *testing.T) {
		f := NewFactory(DefaultConfig())
		f.SetOciStore(store)
		f.ProvisionRuntime(map[string]*config.Runtime{
			"node": {Image: "drassi/node:24"},
		})
		eng, err := f.Create()
		require.NoError(t, err)
		require.NotNil(t, eng)
		_ = eng.Close()
	})

	t.Run("with runtimes but missing store panics", func(t *testing.T) {
		f := NewFactory(DefaultConfig())
		f.ProvisionRuntime(map[string]*config.Runtime{
			"node": {Image: "drassi/node:24"},
		})
		require.Panicsf(t, func() {
			_, _ = f.Create()
		}, "oci store required")
	})
}

func TestNew(t *testing.T) {
	t.Run("with nil config uses DefaultConfig without panicking", func(t *testing.T) {
		eng, err := New(nil, nil)
		require.NoError(t, err)
		require.NotNil(t, eng)
		_ = eng.Close()
	})
}

func TestDockerEngineLifecycleSuite(t *testing.T) {
	suite.Run(t, new(DockerEngineLifecycleTestSuite))
}

type DockerEngineLifecycleTestSuite struct {
	suite.Suite
	ctrl       *gomock.Controller
	mockClient *mock_container.MockEngine
	forge      *records.Forge
}

func (s *DockerEngineLifecycleTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.forge = &records.Forge{
		Repository: "drassi/test",
		Workflow:   "test.yml",
		Job:        "build",
		RunId:      "100",
		RunAttempt: "1",
	}
	s.mockClient = mock_container.NewMockEngine(s.ctrl)
	s.mockClient.EXPECT().
		VolumeCreate(gomock.Any(), gomock.Any()).
		Return(s.forge.CanonicalName(), nil).
		AnyTimes()
}

func (s *DockerEngineLifecycleTestSuite) TestLaunch_Case1_NoJobContainer() {
	s.mockClient.EXPECT().NetworkCreate(gomock.Any(), gomock.Any()).Return("net-docker-1", nil)
	s.mockClient.EXPECT().Address().Return("unix:///var/run/docker.sock")
	s.mockClient.EXPECT().ContainerRun(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, spec *types.ContainerSpec, _ *c.RunOptions) (string, error) {
			s.Require().Equal(container.DefaultImage, spec.Image)
			s.Require().Equal([]string{"sleep"}, spec.Entrypoint)
			s.Require().Equal([]string{"infinity"}, spec.Command)
			s.Require().Len(spec.Endpoints, 1)
			s.Require().Equal("net-docker-1", spec.Endpoints[0].Target)

			var jobVolFound bool
			for _, m := range spec.Mounts {
				if m.Type == "volume" && m.Target == container.DefaultJobDir {
					jobVolFound = true
					s.Require().Equal(s.forge.CanonicalName(), m.Source)
					s.Require().NotNil(m.VolumeOptions)
					s.Require().Equal(s.forge.WellKnownLabels(), m.VolumeOptions.Labels)
				}
			}
			s.Require().True(jobVolFound, "empty volume for job dir must be mounted in docker sandbox")
			return "c-sb-1", nil
		},
	)
	s.mockClient.EXPECT().CopyIn(gomock.Any(), "c-sb-1", gomock.Any()).Return(nil)
	s.mockClient.EXPECT().ContainerInspect(gomock.Any(), "c-sb-1").Return(&types.ContainerSpec{
		Environment: map[string]string{"PATH": "/bin"},
	}, nil)

	eng := &engine{client: s.mockClient, template: DefaultConfig().Template}
	resp, err := eng.Launch(s.T().Context(), &sandboxer.LaunchRequest{Forge: s.forge})
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().NotNil(resp.Sandbox)
	s.Require().NotNil(resp.JobContainer)
	s.Require().Equal("c-sb-1", resp.JobContainer.Id)
	s.Require().NotNil(resp.Mounter)

	s.Require().NotNil(resp.ContainerEngine)
	ce, err := resp.ContainerEngine(s.T().Context())
	s.Require().NoError(err)
	s.Require().Equal(s.mockClient, ce)

	s.mockClient.EXPECT().ContainerRemove(gomock.Any(), &c.RemoveOptions{Id: "c-sb-1"}).Return(nil)
	s.mockClient.EXPECT().ContainerRemove(gomock.Any(), &c.RemoveOptions{Labels: s.forge.WellKnownLabels()}).Return(nil)
	s.mockClient.EXPECT().VolumeRemove(gomock.Any(), &c.RemoveOptions{Labels: s.forge.WellKnownLabels()}).Return(nil)
	s.mockClient.EXPECT().NetworkRemove(gomock.Any(), &c.RemoveOptions{Id: "net-docker-1"}).Return(nil)
	s.Require().NoError(resp.Sandbox.Terminate(s.T().Context()))
}

func (s *DockerEngineLifecycleTestSuite) TestLaunch_Case2_WithJobContainer() {
	s.mockClient.EXPECT().NetworkCreate(gomock.Any(), gomock.Any()).Return("net-docker-2", nil)
	s.mockClient.EXPECT().Address().Return("unix:///var/run/docker.sock")
	s.mockClient.EXPECT().ContainerRun(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, spec *types.ContainerSpec, _ *c.RunOptions) (string, error) {
			s.Require().Equal("node:18", spec.Image)
			s.Require().Equal("custom-val", spec.Environment["CUSTOM_ENV"])

			var jobVolFound bool
			for _, m := range spec.Mounts {
				if m.Type == "volume" && m.Target == container.DefaultJobDir {
					jobVolFound = true
					s.Require().Equal(s.forge.CanonicalName(), m.Source)
					s.Require().NotNil(m.VolumeOptions)
					s.Require().Equal(s.forge.WellKnownLabels(), m.VolumeOptions.Labels)
				}
			}
			s.Require().True(jobVolFound, "empty volume for job dir must be mounted in docker sandbox")
			return "c-node-1", nil
		},
	)
	s.mockClient.EXPECT().CopyIn(gomock.Any(), "c-node-1", gomock.Any()).Return(nil)
	s.mockClient.EXPECT().ContainerInspect(gomock.Any(), "c-node-1").Return(&types.ContainerSpec{
		Environment: map[string]string{"PATH": "/bin"},
	}, nil)

	eng := &engine{client: s.mockClient, template: DefaultConfig().Template}
	req := &sandboxer.LaunchRequest{
		Forge: s.forge,
		JobContainer: &workflows.Container{
			Image: "node:18",
			Env:   map[string]string{"CUSTOM_ENV": "custom-val"},
		},
	}
	resp, err := eng.Launch(s.T().Context(), req)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Equal("c-node-1", resp.JobContainer.Id)

	s.mockClient.EXPECT().ContainerRemove(gomock.Any(), &c.RemoveOptions{Id: "c-node-1"}).Return(nil)
	s.mockClient.EXPECT().ContainerRemove(gomock.Any(), &c.RemoveOptions{Labels: s.forge.WellKnownLabels()}).Return(nil)
	s.mockClient.EXPECT().VolumeRemove(gomock.Any(), &c.RemoveOptions{Labels: s.forge.WellKnownLabels()}).Return(nil)
	s.mockClient.EXPECT().NetworkRemove(gomock.Any(), &c.RemoveOptions{Id: "net-docker-2"}).Return(nil)
	s.Require().NoError(resp.Sandbox.Terminate(s.T().Context()))
}

func (s *DockerEngineLifecycleTestSuite) TestLaunch_Case2b_WithServices() {
	s.mockClient.EXPECT().NetworkCreate(gomock.Any(), gomock.Any()).Return("net-docker-3", nil)
	s.mockClient.EXPECT().Address().Return("unix:///var/run/docker.sock")
	s.mockClient.EXPECT().ContainerRun(gomock.Any(), gomock.Any(), gomock.Any()).Return("c-main", nil)
	s.mockClient.EXPECT().CopyIn(gomock.Any(), "c-main", gomock.Any()).Return(nil)
	s.mockClient.EXPECT().ContainerInspect(gomock.Any(), "c-main").Return(&types.ContainerSpec{
		Environment: map[string]string{"PATH": "/bin"},
	}, nil)

	// Bootstrap for service container
	s.mockClient.EXPECT().ImagePull(gomock.Any(), "redis:7", gomock.Any()).Return(nil)
	s.mockClient.EXPECT().ContainerRun(gomock.Any(), gomock.Any(), gomock.Any()).Return("c-redis-svc", nil)
	s.mockClient.EXPECT().ContainerInspect(gomock.Any(), "c-redis-svc").Return(&types.ContainerSpec{}, nil)

	eng := &engine{client: s.mockClient, template: DefaultConfig().Template}
	req := &sandboxer.LaunchRequest{
		Forge: s.forge,
		ServiceContainers: map[string]*workflows.Container{
			"redis": {Image: "redis:7"},
		},
	}
	resp, err := eng.Launch(s.T().Context(), req)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Contains(resp.ServiceContainers, "redis")
	s.Require().Equal("c-redis-svc", resp.ServiceContainers["redis"].Id)

	s.mockClient.EXPECT().ContainerRemove(gomock.Any(), &c.RemoveOptions{Id: "c-main"}).Return(nil)
	s.mockClient.EXPECT().ContainerRemove(gomock.Any(), &c.RemoveOptions{Labels: s.forge.WellKnownLabels()}).Return(nil)
	s.mockClient.EXPECT().VolumeRemove(gomock.Any(), &c.RemoveOptions{Labels: s.forge.WellKnownLabels()}).Return(nil)
	s.mockClient.EXPECT().NetworkRemove(gomock.Any(), &c.RemoveOptions{Id: "net-docker-3"}).Return(nil)
	s.Require().NoError(resp.Sandbox.Terminate(s.T().Context()))
}

func (s *DockerEngineLifecycleTestSuite) TestLaunch_FailureRollback() {
	s.mockClient.EXPECT().NetworkCreate(gomock.Any(), gomock.Any()).Return("net-fail", nil)
	s.mockClient.EXPECT().Address().Return("unix:///var/run/docker.sock")
	s.mockClient.EXPECT().ContainerRun(gomock.Any(), gomock.Any(), gomock.Any()).Return("", errors.New("run failed"))
	s.mockClient.EXPECT().ContainerRemove(gomock.Any(), &c.RemoveOptions{Labels: s.forge.WellKnownLabels()}).Return(nil)
	s.mockClient.EXPECT().VolumeRemove(gomock.Any(), &c.RemoveOptions{Labels: s.forge.WellKnownLabels()}).Return(nil)
	s.mockClient.EXPECT().NetworkRemove(gomock.Any(), &c.RemoveOptions{Id: "net-fail"}).Return(nil)

	eng := &engine{client: s.mockClient, template: DefaultConfig().Template}
	resp, err := eng.Launch(s.T().Context(), &sandboxer.LaunchRequest{Forge: s.forge})
	s.Require().Error(err)
	s.Require().Nil(resp)
	s.Require().Contains(err.Error(), "run failed")
}
