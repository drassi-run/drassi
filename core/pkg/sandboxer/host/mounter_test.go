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

	mock_sandboxer "drassi.run/core/mock/sandboxer"
	"drassi.run/core/pkg/sandboxer"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

func TestHostMounterSuite(t *testing.T) {
	suite.Run(t, new(HostMounterTestSuite))
}

type HostMounterTestSuite struct {
	suite.Suite
}

func (s *HostMounterTestSuite) TestOverlayMounts() {
	s.Run("coarse=true with StandardLayout returns whole job directory", func() {
		tempDir := s.T().TempDir()
		sandboxDir := filepath.Join(tempDir, "job")
		s.Require().NoError(os.MkdirAll(sandboxDir, 0755))

		m := newMounter(sandboxDir)
		target := sandboxer.StandardLayout("/opt/drassi/")

		mounts := m.OverlayMounts(target, true)
		s.NotEmpty(mounts)
		s.Equal(sandboxDir, mounts[0].Source)
		s.Equal(string(target), mounts[0].Target)
	})

	s.Run("coarse=false returns discrete folders and resolves runtimes symlinks", func() {
		tempDir := s.T().TempDir()
		sandboxDir := filepath.Join(tempDir, "job")
		runtimesDir := filepath.Join(sandboxDir, "runtimes")
		s.Require().NoError(os.MkdirAll(runtimesDir, 0755))

		// Create a real OCI cache directory and symlink it
		ociDir := filepath.Join(tempDir, "oci-cache", "node-18")
		s.Require().NoError(os.MkdirAll(ociDir, 0755))
		symlinkNode := filepath.Join(runtimesDir, "node")
		s.Require().NoError(os.Symlink(ociDir, symlinkNode))

		layout := sandboxer.StandardLayout(sandboxDir)
		m := newMounter(sandboxDir)
		target := sandboxer.StandardLayout("/opt/drassi/")

		mounts := m.OverlayMounts(target, false)

		// Must contain workspace, temp, actions, tools, and the resolved node runtime mount
		var nodeMountFound bool
		targets := make(map[string]string)
		for _, mount := range mounts {
			targets[mount.Target] = mount.Source
			if mount.Target == "/opt/drassi/runtimes/node" {
				nodeMountFound = true
				s.Equal(ociDir, mount.Source, "Source must be the resolved real host path, not the symlink")
				s.True(mount.ReadOnly)
			}
		}
		s.True(nodeMountFound, "Runtimes mount MUST be passed to job container")
		s.Equal(layout.Workspace(), targets["/opt/drassi/workspace"])
		s.Equal(layout.Temp(), targets["/opt/drassi/temp"])
		s.Equal(layout.Actions(), targets["/opt/drassi/actions"])
		s.Equal(layout.Tools(), targets["/opt/drassi/tools"])
	})

	s.Run("coarse=true with non-standard layout falls back to discrete folders", func() {
		tempDir := s.T().TempDir()
		sandboxDir := filepath.Join(tempDir, "job")
		s.Require().NoError(os.MkdirAll(sandboxDir, 0755))

		layout := sandboxer.StandardLayout(sandboxDir)
		m := newMounter(sandboxDir)
		target := newMockLayout(gomock.NewController(s.T()))

		mounts := m.OverlayMounts(target, true)
		s.Len(mounts, 4)
		targets := make(map[string]string)
		for _, mount := range mounts {
			targets[mount.Target] = mount.Source
		}
		s.Equal(layout.Workspace(), targets["/custom/ws"])
		s.Equal(layout.Temp(), targets["/custom/tmp"])
	})

	s.Run("broken symlink skipped", func() {
		tempDir := s.T().TempDir()
		sandboxDir := filepath.Join(tempDir, "job")
		runtimesDir := filepath.Join(sandboxDir, "runtimes")
		s.Require().NoError(os.MkdirAll(runtimesDir, 0755))

		// Create a broken symlink (points to nonexistent path)
		brokenSymlink := filepath.Join(runtimesDir, "broken")
		s.Require().NoError(os.Symlink(filepath.Join(tempDir, "nonexistent"), brokenSymlink))

		m := newMounter(sandboxDir)
		target := sandboxer.StandardLayout("/opt/drassi/")
		mounts := m.OverlayMounts(target, false)
		for _, mount := range mounts {
			s.NotEqual("/opt/drassi/runtimes/broken", mount.Target)
		}
	})
}

func newMockLayout(ctrl *gomock.Controller) sandboxer.Layout {
	target := mock_sandboxer.NewMockLayout(ctrl)
	target.EXPECT().Workspace().Return("/custom/ws")
	target.EXPECT().Temp().Return("/custom/tmp")
	target.EXPECT().Actions().Return("/custom/act")
	target.EXPECT().Tools().Return("/custom/tools")
	return target
}

func (s *HostMounterTestSuite) TestRuntimeMounts() {
	s.Run("workspace and temp mounts", func() {
		tempDir := s.T().TempDir()
		sandboxDir := filepath.Join(tempDir, "job")

		layout := sandboxer.StandardLayout(sandboxDir)
		m := newMounter(sandboxDir)
		target := sandboxer.StandardLayout("/opt/drassi/")
		mounts := m.RuntimeMounts(target)
		s.Require().Len(mounts, 2)
		s.Equal(layout.Workspace(), mounts[0].Source)
		s.Equal(target.Workspace(), mounts[0].Target)
		s.Equal(layout.Temp(), mounts[1].Source)
		s.Equal(target.Temp(), mounts[1].Target)
	})
}
