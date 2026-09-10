/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package host

import (
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
	ctrl       *gomock.Controller
	store      *mock_store.MockManager
	tempDir    string
	runtimeDir string
	cfg        *Config
}

func (s *HostEngineTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.store = mock_store.NewMockManager(s.ctrl)
	s.tempDir = s.T().TempDir()
	s.runtimeDir = filepath.Join(s.tempDir, "opt_drassi_runtimes")
	s.cfg = &Config{
		RootDir:    s.tempDir,
		RuntimeDir: s.runtimeDir,
	}
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
	return resp.Sandbox
}

func (s *HostEngineTestSuite) TestLaunch() {
	s.Run("with provisioner", func() {
		mountDir := filepath.Join(s.tempDir, "node_mount")
		s.Require().NoError(os.MkdirAll(mountDir, 0755))

		img := &ocistore.Image{}
		s.store.EXPECT().Image(gomock.Any(), "drassi/node:24").Return(img, nil).Times(1)
		s.store.EXPECT().Mount(gomock.Any(), img, gomock.Any()).Return(mountDir, "layer-node", nil).Times(1)
		s.store.EXPECT().Unmount(gomock.Any(), "layer-node").Return(nil).Times(1)

		runtimes := map[string]*config.Runtime{
			"node": {Image: "drassi/node:24"},
		}

		p := provision.New[string](
			runtimes,
			provision.Pull[string](s.store),
			provision.Mount[string](s.store),
			Symlink[string](),
		)

		eng, err := New(s.cfg, p)
		s.Require().NoError(err)

		sb := s.assertLaunch(eng)

		// Symlink is created
		symlinkPath := filepath.Join(s.runtimeDir, "node")
		target, err := os.Readlink(symlinkPath)
		s.Require().NoError(err)
		s.Require().Equal(mountDir, target)

		// Terminate sandbox unmounts layers and removes workspace
		s.Require().NoError(sb.Terminate(s.T().Context()))
	})

	s.Run("without provisioner", func() {
		eng, err := New(s.cfg, nil)
		s.Require().NoError(err)

		sb := s.assertLaunch(eng)
		s.Require().NoError(sb.Terminate(s.T().Context()))
	})
}


func (s *HostEngineTestSuite) TestFactory() {
	s.Run("with runtimes and store", func() {
		cfg := DefaultConfig()
		cfg.RootDir = s.T().TempDir()
		cfg.RuntimeDir = filepath.Join(cfg.RootDir, "runtimes")
		f := NewFactory(cfg)
		f.SetOciStore(s.store)
		f.ProvisionRuntime(map[string]*config.Runtime{
			"node": {Image: "drassi/node:24"},
		})
		eng, err := f.Create()
		s.Require().NoError(err)
		s.Require().NotNil(eng)
		_ = eng.Close()
	})

	s.Run("with runtimes but missing store returns error", func() {
		cfg := DefaultConfig()
		cfg.RootDir = s.T().TempDir()
		cfg.RuntimeDir = filepath.Join(cfg.RootDir, "runtimes")
		f := NewFactory(cfg)
		f.ProvisionRuntime(map[string]*config.Runtime{
			"node": {Image: "drassi/node:24"},
		})
		_, err := f.Create()
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "oci store is required")
	})

	s.Run("without runtimes", func() {
		cfg := DefaultConfig()
		cfg.RootDir = s.T().TempDir()
		cfg.RuntimeDir = filepath.Join(cfg.RootDir, "runtimes")
		f := NewFactory(cfg)
		eng, err := f.Create()
		s.Require().NoError(err)
		s.Require().NotNil(eng)
		_ = eng.Close()
	})
}

func (s *HostEngineTestSuite) TestNew() {
	s.Run("without panic when nil", func() {
		cfg := DefaultConfig()
		cfg.RootDir = s.T().TempDir()
		eng, err := New(cfg, nil)
		s.Require().NoError(err)
		s.Require().NotNil(eng)
		_ = eng.Close()
	})
}
