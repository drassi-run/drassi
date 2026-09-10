/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package provision_test

import (
	"context"
	"testing"
	"time"

	"drassi.run/core/config"
	"drassi.run/core/pkg/runtime/provision"
	ocistore "drassi.run/core/pkg/store/oci"
	"github.com/stretchr/testify/suite"
)

func TestContextSuite(t *testing.T) {
	suite.Run(t, new(ContextTestSuite))
}

type ContextTestSuite struct {
	suite.Suite
}

func (s *ContextTestSuite) TestGenericState() {
	ctx := s.T().Context()
	rtCfg := &config.Runtime{Image: "drassi/node:24"}
	pctx := provision.NewContext(ctx, "node", rtCfg, "/opt/drassi/runtimes/node")

	s.Require().Equal("node", pctx.RuntimeName)
	s.Require().Equal("/opt/drassi/runtimes/node", pctx.TargetDir)
	s.Require().Equal(rtCfg, pctx.Config)

	// Test typed key string
	pctx.Set(provision.KeyHostMountDir, "/var/lib/drassi/storage/overlay/merged")
	val, ok := pctx.Get(provision.KeyHostMountDir)
	s.Require().True(ok)
	s.Require().Equal("/var/lib/drassi/storage/overlay/merged", val)
	s.Require().Equal("/var/lib/drassi/storage/overlay/merged", pctx.MustGet(provision.KeyHostMountDir))

	// Test KeyMountID
	pctx.Set(provision.KeyMountID, "mount-12345")
	mountID, ok := pctx.Get(provision.KeyMountID)
	s.Require().True(ok)
	s.Require().Equal("mount-12345", mountID)
	s.Require().Equal("mount-12345", pctx.MustGet(provision.KeyMountID))

	// Test KeyImage
	img := &ocistore.Image{ID: "img-node"}
	pctx.Set(provision.KeyImage, img)
	gotImg, ok := pctx.Get(provision.KeyImage)
	s.Require().True(ok)
	s.Require().Equal(img, gotImg)
	s.Require().Equal(img, pctx.MustGet(provision.KeyImage))

	// Test missing key
	const keyMissing = provision.StateKey[int]("missing_key")
	intVal, ok := pctx.Get(keyMissing)
	s.Require().False(ok)
	s.Require().Equal(0, intVal)
	s.Require().Panics(func() {
		pctx.MustGet(keyMissing)
	})

	// Test type mismatch
	const keyMismatch = provision.StateKey[int]("host_mount_dir")
	mismatchVal, ok := pctx.Get(keyMismatch)
	s.Require().False(ok)
	s.Require().Equal(0, mismatchVal)
	s.Require().Panics(func() {
		pctx.MustGet(keyMismatch)
	})
}

func (s *ContextTestSuite) TestCancellation() {
	ctx, cancel := context.WithCancel(s.T().Context())
	pctx := provision.NewContext(ctx, "node", nil, "")

	select {
	case <-pctx.Done():
		s.T().Fatal("context should not be done yet")
	default:
	}

	cancel()

	select {
	case <-pctx.Done():
		s.Require().Equal(context.Canceled, pctx.Err())
	case <-time.After(time.Second):
		s.T().Fatal("timed out waiting for context cancellation")
	}
}
