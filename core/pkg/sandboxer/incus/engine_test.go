/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package incus

import (
	"context"
	"testing"

	"drassi.run/core/config"
	mock_sandboxer "drassi.run/core/mock/sandboxer"
	mock_store "drassi.run/core/mock/store/oci"
	"drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/sandboxer"
	ocistore "drassi.run/core/pkg/store/oci"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

func TestIncusEngineSuite(t *testing.T) {
	suite.Run(t, new(IncusEngineTestSuite))
}

type IncusEngineTestSuite struct {
	suite.Suite
	ctrl  *gomock.Controller
	store *mock_store.MockManager
}

func (s *IncusEngineTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.store = mock_store.NewMockManager(s.ctrl)
}

func (s *IncusEngineTestSuite) TestProvisionerIntegration() {
	img := &ocistore.Image{}
	s.store.EXPECT().Image(gomock.Any(), "drassi/node:24").Return(img, nil).Times(1)
	s.store.EXPECT().Mount(gomock.Any(), img, gomock.Any()).Return("/var/lib/drassi/node_mount", "layer-node", nil).Times(1)
	s.store.EXPECT().Unmount(gomock.Any(), "layer-node").Return(nil).Times(1)

	runtimes := map[string]*config.Runtime{
		"node": {Image: "drassi/node:24"},
	}

	p := provision.New[*Template](
		runtimes,
		provision.Pull[*Template](s.store),
		provision.Mount[*Template](s.store),
		AddDiskDevice(),
	)

	tmpl := &Template{}
	launched := false
	mockSb := mock_sandboxer.NewMockSandbox(s.ctrl)
	mockSb.EXPECT().Terminate(gomock.Any()).Return(nil).Times(1)

	launcher := func(ctx context.Context, tmpl *Template) (sandboxer.Sandbox, error) {
		launched = true
		s.Require().Contains(tmpl.Devices, "runtime-node")
		s.Require().Equal("disk", tmpl.Devices["runtime-node"]["type"])
		s.Require().Equal("/var/lib/drassi/node_mount", tmpl.Devices["runtime-node"]["source"])
		s.Require().Equal("/opt/drassi/runtimes/node", tmpl.Devices["runtime-node"]["path"])
		return mockSb, nil
	}

	sb, err := p.Launch("/opt/drassi/runtimes", launcher)(s.T().Context(), tmpl)
	s.Require().NoError(err)
	s.Require().True(launched)
	s.Require().NotNil(sb)

	// Verify unmount cleanup runs on terminate
	s.Require().NoError(sb.Terminate(s.T().Context()))
}

func (s *IncusEngineTestSuite) TestTemplateClone() {
	s.Run("nil template", func() {
		var tmpl *Template
		s.Require().Nil(tmpl.Clone())
	})

	s.Run("deep copy fields", func() {
		orig := &Template{
			Name:         "test-instance",
			Image:        "ubuntu:22.04",
			Architecture: "x86_64",
			InstanceSize: "t1.micro",
			Profiles:     []string{"default", "custom"},
			Config: map[string]string{
				"security.nesting": "true",
			},
			Devices: map[string]map[string]string{
				"root": {
					"type": "disk",
					"path": "/",
				},
			},
			Ephemeral: true,
		}

		clone := orig.Clone()
		s.Require().NotNil(clone)
		s.Require().Equal(orig.Name, clone.Name)
		s.Require().Equal(orig.Image, clone.Image)
		s.Require().Equal(orig.Architecture, clone.Architecture)
		s.Require().Equal(orig.InstanceSize, clone.InstanceSize)
		s.Require().Equal(orig.Ephemeral, clone.Ephemeral)
		s.Require().Equal(orig.Profiles, clone.Profiles)
		s.Require().Equal(orig.Config, clone.Config)
		s.Require().Equal(orig.Devices, clone.Devices)

		// Verify Profiles is deep copied
		clone.Profiles[0] = "modified"
		s.Require().Equal("default", orig.Profiles[0])

		// Verify Config is deep copied
		clone.Config["security.nesting"] = "false"
		clone.Config["new.key"] = "val"
		s.Require().Equal("true", orig.Config["security.nesting"])
		s.Require().NotContains(orig.Config, "new.key")

		// Verify Devices outer map is deep copied
		clone.Devices["extra"] = map[string]string{"type": "nic"}
		s.Require().NotContains(orig.Devices, "extra")

		// Verify Devices inner map is deep copied
		clone.Devices["root"]["path"] = "/mnt"
		s.Require().Equal("/", orig.Devices["root"]["path"])
	})
}

func (s *IncusEngineTestSuite) TestFactory() {
	s.Run("with runtimes and store", func() {
		f := NewFactory(DefaultConfig())
		f.SetOciStore(s.store)
		f.ProvisionRuntime(map[string]*config.Runtime{
			"node": {Image: "drassi/node:24"},
		})
		s.Require().NotPanics(func() {
			_, _ = f.Create()
		})
	})

	s.Run("with runtimes but missing store returns error", func() {
		f := NewFactory(DefaultConfig())
		f.ProvisionRuntime(map[string]*config.Runtime{
			"node": {Image: "drassi/node:24"},
		})
		_, err := f.Create()
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "oci store is required")
	})

	s.Run("without runtimes", func() {
		f := NewFactory(DefaultConfig())
		s.Require().NotPanics(func() {
			_, _ = f.Create()
		})
	})
}

func (s *IncusEngineTestSuite) TestNew() {
	s.Run("without panic when nil", func() {
		cfg := DefaultConfig()
		s.Require().NotPanics(func() {
			_, _ = New(cfg, nil)
		})
	})
}
