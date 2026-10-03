/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package provision

import (
	"context"
	"testing"
	"time"

	"drassi.run/core/config"
	mock_store "drassi.run/core/mock/store/oci"
	ocistore "drassi.run/core/pkg/store/oci"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
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
	pctx := NewContext(ctx, "node", rtCfg, "/opt/drassi/runtimes/node")

	s.Require().Equal("node", pctx.RuntimeName)
	s.Require().Equal("/opt/drassi/runtimes/node", pctx.TargetDir)
	s.Require().Equal(rtCfg, pctx.Config)

	// Test typed key string
	pctx.Set(KeyHostMountDir, "/var/lib/drassi/storage/overlay/merged")
	val, ok := pctx.Get(KeyHostMountDir)
	s.Require().True(ok)
	s.Require().Equal("/var/lib/drassi/storage/overlay/merged", val)
	s.Require().Equal("/var/lib/drassi/storage/overlay/merged", pctx.MustGet(KeyHostMountDir))

	// Test KeyImage
	var img ocistore.Image = mock_store.NewMockImage(gomock.NewController(s.T()))
	pctx.Set(KeyImage, img)
	gotImg, ok := pctx.Get(KeyImage)
	s.Require().True(ok)
	s.Require().Equal(img, gotImg)
	s.Require().Equal(img, pctx.MustGet(KeyImage))

	// Test missing key
	const keyMissing = StateKey[int]("missing_key")
	intVal, ok := pctx.Get(keyMissing)
	s.Require().False(ok)
	s.Require().Equal(0, intVal)
	s.Require().Panics(func() {
		pctx.MustGet(keyMissing)
	})

	// Test type mismatch
	const keyMismatch = StateKey[int]("host_mount_dir")
	mismatchVal, ok := pctx.Get(keyMismatch)
	s.Require().False(ok)
	s.Require().Equal(0, mismatchVal)
	s.Require().Panics(func() {
		pctx.MustGet(keyMismatch)
	})
}

func (s *ContextTestSuite) TestCancellation() {
	ctx, cancel := context.WithCancel(s.T().Context())
	pctx := NewContext(ctx, "node", nil, "")

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
