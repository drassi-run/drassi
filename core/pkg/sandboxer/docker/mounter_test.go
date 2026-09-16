/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package docker

import (
	"testing"

	mock_sandboxer "drassi.run/core/mock/sandboxer"
	"drassi.run/core/pkg/sandboxer"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

func TestDockerMounterSuite(t *testing.T) {
	suite.Run(t, new(DockerMounterTestSuite))
}

type DockerMounterTestSuite struct {
	suite.Suite
}

func (s *DockerMounterTestSuite) TestOverlayMounts() {
	m := newMounter("shared-vol")
	target := mock_sandboxer.NewMockLayout(gomock.NewController(s.T()))
	mounts := m.OverlayMounts(target, false)
	s.Require().Len(mounts, 0)
}

func (s *DockerMounterTestSuite) TestRuntimeMounts() {
	s.Run("returns workspace and temp volume mounts with subpath", func() {
		m := newMounter("shared-vol")
		s.Require().NotNil(m)

		target := sandboxer.StandardLayout("/github")
		mounts := m.RuntimeMounts(target)
		s.Require().Len(mounts, 2)

		s.Equal("volume", mounts[0].Type)
		s.Equal("shared-vol", mounts[0].Source)
		s.Equal("/github/workspace", mounts[0].Target)
		s.Require().NotNil(mounts[0].VolumeOptions)
		s.Equal("workspace", mounts[0].VolumeOptions.SubPath)

		s.Equal("volume", mounts[1].Type)
		s.Equal("shared-vol", mounts[1].Source)
		s.Equal("/github/temp", mounts[1].Target)
		s.Require().NotNil(mounts[1].VolumeOptions)
		s.Equal("temp", mounts[1].VolumeOptions.SubPath)
	})
}
