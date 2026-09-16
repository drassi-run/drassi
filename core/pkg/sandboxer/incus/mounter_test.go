/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package incus

import (
	"testing"

	mock_sandboxer "drassi.run/core/mock/sandboxer"
	"drassi.run/core/pkg/sandboxer"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

func TestIncusMounterSuite(t *testing.T) {
	suite.Run(t, new(IncusMounterTestSuite))
}

type IncusMounterTestSuite struct {
	suite.Suite
}

func (s *IncusMounterTestSuite) TestOverlayMounts() {
	s.Run("coarse=true with StandardLayout returns whole job directory", func() {
		m := newMounter()
		target := sandboxer.StandardLayout("/opt/drassi/")
		mounts := m.OverlayMounts(target, true)
		s.Len(mounts, 1)
		s.Equal("bind", mounts[0].Type)
		s.Equal(string(layout), mounts[0].Source)
		s.Equal(string(target), mounts[0].Target)
		s.NotNil(mounts[0].BindOptions)
		s.Equal("enabled", mounts[0].BindOptions.Recursive)
	})

	s.Run("coarse=false returns individual folders", func() {
		m := newMounter()
		target := sandboxer.StandardLayout("/opt/drassi/")
		mounts := m.OverlayMounts(target, false)
		s.Len(mounts, 5)

		targets := make(map[string]string)
		for _, mount := range mounts {
			targets[mount.Target] = mount.Source
			if mount.Target == "/opt/drassi/runtimes" {
				s.NotNil(mount.BindOptions)
				s.Equal("enabled", mount.BindOptions.Recursive)
			}
		}
		s.Equal("/opt/drassi/workspace", targets["/opt/drassi/workspace"])
		s.Equal("/opt/drassi/temp", targets["/opt/drassi/temp"])
		s.Equal("/opt/drassi/actions", targets["/opt/drassi/actions"])
		s.Equal("/opt/drassi/tools", targets["/opt/drassi/tools"])
		s.Equal("/opt/drassi/runtimes", targets["/opt/drassi/runtimes"])
	})

	s.Run("coarse=true with non-standard layout falls back to individual folders", func() {
		m := newMounter()
		target := newMockLayout(gomock.NewController(s.T()))

		mounts := m.OverlayMounts(target, true)
		s.Len(mounts, 5)
		targets := make(map[string]string)
		for _, mount := range mounts {
			targets[mount.Target] = mount.Source
		}
		s.Equal("/opt/drassi/workspace", targets["/custom/ws"])
		s.Equal("/opt/drassi/temp", targets["/custom/tmp"])
	})
}

func newMockLayout(ctrl *gomock.Controller) sandboxer.Layout {
	target := mock_sandboxer.NewMockLayout(ctrl)
	target.EXPECT().Workspace().Return("/custom/ws")
	target.EXPECT().Temp().Return("/custom/tmp")
	target.EXPECT().Actions().Return("/custom/act")
	target.EXPECT().Tools().Return("/custom/tools")
	target.EXPECT().Runtimes().Return("/custom/run")
	return target
}

func (s *IncusMounterTestSuite) TestRuntimeMounts() {
	s.Run("valid target layout", func() {
		m := newMounter()
		actionTarget := sandboxer.StandardLayout("/opt/drassi/")

		mounts := m.RuntimeMounts(actionTarget)
		s.Len(mounts, 2)

		targets := make(map[string]string)
		for _, mount := range mounts {
			s.Equal("bind", mount.Type)
			targets[mount.Target] = mount.Source
		}
		s.Equal("/opt/drassi/workspace", targets["/opt/drassi/workspace"])
		s.Equal("/opt/drassi/temp", targets["/opt/drassi/temp"])
	})
}
