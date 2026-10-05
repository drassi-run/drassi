/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package host

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"drassi.run/core/config"
	mock_store "drassi.run/core/mock/store/oci"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/sandboxer"
	ocistore "drassi.run/core/pkg/store/oci"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

func TestHostEngineSuite(t *testing.T) {
	suite.Run(t, new(HostEngineTestSuite))
}

type HostEngineTestSuite struct {
	suite.Suite
	ctrl    *gomock.Controller
	store   *mock_store.MockManager
	tempDir string
}

func (s *HostEngineTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.store = mock_store.NewMockManager(s.ctrl)
	s.tempDir = s.T().TempDir()
}

func (s *HostEngineTestSuite) assertLaunch(eng sandboxer.Engine) sandboxer.Sandbox {
	req := &sandboxer.LaunchRequest{
		Forge: &records.Forge{
			Repository: "drassi/test",
			Workflow:   "build.yml",
			Job:        "test",
			RunId:      "1",
			RunAttempt: "1",
		},
	}

	resp, err := eng.Launch(s.T().Context(), req)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().NotNil(resp.Sandbox)
	s.Require().NotNil(resp.Mounter)
	return resp.Sandbox
}

func (s *HostEngineTestSuite) TestLaunch() {
	s.Run("with provisioner", func() {
		mountDir := filepath.Join(s.tempDir, "drassi/test/build/test/1_1/runtimes/node")
		s.Require().NoError(os.MkdirAll(mountDir, 0755))

		img := mock_store.NewMockImage(s.ctrl)
		released := false
		mnt := &ocistore.Mount{
			Target: mountDir,
			Release: func(context.Context) error {
				released = true
				return nil
			},
		}

		s.store.EXPECT().Pull(gomock.Any(), "drassi/node:24").Return(img, nil).Times(1)
		s.store.EXPECT().Mount(gomock.Any(), img, gomock.Any()).Return(mnt, nil).Times(1)

		runtimes := map[string]*config.Runtime{
			"node": {Image: "drassi/node:24"},
		}

		fact := func(runtimeDir string, name string, rt config.Runtime) provision.Pipeline[string] {
			targetDir := filepath.Join(runtimeDir, name)
			state := new(provision.State)
			return provision.Pipeline[string]{
				provision.Pull[string](s.store, state, rt.Image),
				provision.Mount[string](s.store, state,
					ocistore.WithTarget(targetDir),
					ocistore.WithWritable(!rt.ReadOnly),
				),
			}
		}

		p := provision.New[string](runtimes, fact)

		eng, err := New(s.tempDir, p)
		s.Require().NoError(err)

		sb := s.assertLaunch(eng)

		// Terminate sandbox unmounts layers and removes workspace
		s.Require().NoError(sb.Terminate(s.T().Context()))
		s.Require().True(released)
	})

	s.Run("without provisioner", func() {
		eng, err := New(s.tempDir, nil)
		s.Require().NoError(err)

		sb := s.assertLaunch(eng)
		s.Require().NoError(sb.Terminate(s.T().Context()))
	})
}

func (s *HostEngineTestSuite) TestLaunch_WithoutContainers_NoDocker() {
	eng, err := New(s.tempDir, nil)
	s.Require().NoError(err)

	req := &sandboxer.LaunchRequest{
		Forge: &records.Forge{
			Repository: "drassi/test",
			Workflow:   "build.yml",
			Job:        "test",
			RunId:      "1",
			RunAttempt: "1",
		},
	}

	resp, err := eng.Launch(s.T().Context(), req)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().NotNil(resp.Sandbox)
	s.Require().NotNil(resp.Mounter)
	s.Require().NotNil(resp.ContainerEngine, "ContainerEngine provider must be provided for lazy init")
	s.Require().Nil(resp.JobContainer)
	s.Require().Empty(resp.ServiceContainers)
	s.Require().NoError(resp.Sandbox.Terminate(s.T().Context()))
}

func (s *HostEngineTestSuite) TestFactory() {
	s.Run("with runtimes and store", func() {
		cfg := DefaultConfig()
		f := NewFactory(cfg)
		f.RootDir(s.T().TempDir())
		f.SetOciStore(s.store)
		f.ProvisionRuntime(map[string]*config.Runtime{
			"node": {Image: "drassi/node:24"},
		})
		eng, err := f.Create()
		s.Require().NoError(err)
		s.Require().NotNil(eng)
		_ = eng.Close()
	})

	s.Run("with runtimes but missing store error", func() {
		cfg := DefaultConfig()
		f := NewFactory(cfg)
		f.RootDir(s.T().TempDir())
		f.ProvisionRuntime(map[string]*config.Runtime{
			"node": {Image: "drassi/node:24"},
		})
		_, err := f.Create()
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "oci store required for runtimes")
	})

	s.Run("without runtimes", func() {
		cfg := DefaultConfig()
		f := NewFactory(cfg)
		f.RootDir(s.T().TempDir())
		eng, err := f.Create()
		s.Require().NoError(err)
		s.Require().NotNil(eng)
		_ = eng.Close()
	})
}

func (s *HostEngineTestSuite) TestNew() {
	s.Run("without panic when nil", func() {
		eng, err := New(s.T().TempDir(), nil)
		s.Require().NoError(err)
		s.Require().NotNil(eng)
		_ = eng.Close()
	})
}
