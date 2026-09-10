/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package provision_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"drassi.run/core/config"
	mock_provision "drassi.run/core/mock/runtime/provision"
	mock_sandboxer "drassi.run/core/mock/sandboxer"
	"drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/sandboxer"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

func TestProvisionerSuite(t *testing.T) {
	suite.Run(t, new(ProvisionerTestSuite))
}

type testRequest struct {
	Items []string
}

type ProvisionerTestSuite struct {
	suite.Suite
	ctrl   *gomock.Controller
	baseSb *mock_sandboxer.MockSandbox
	op     *mock_provision.MockOperation[*testRequest]
}

func (s *ProvisionerTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.baseSb = mock_sandboxer.NewMockSandbox(s.ctrl)
	s.op = mock_provision.NewMockOperation[*testRequest](s.ctrl)
}

func (s *ProvisionerTestSuite) TestSuccess() {
	var cleanedUp atomic.Bool
	cleanupFn := func(context.Context) error {
		cleanedUp.Store(true)
		return nil
	}

	s.op.EXPECT().Prepare(gomock.Any()).DoAndReturn(func(pctx *provision.Context) (sandboxer.Cleanup, error) {
		return cleanupFn, nil
	}).Times(2)
	s.op.EXPECT().PreLaunch(gomock.Any(), gomock.Any()).DoAndReturn(func(pctx *provision.Context, req *testRequest) (*testRequest, error) {
		s.Require().Equal("/opt/runtimes/"+pctx.RuntimeName, pctx.TargetDir)
		req.Items = append(req.Items, pctx.RuntimeName)
		return req, nil
	}).Times(2)
	s.op.EXPECT().PostLaunch(gomock.Any(), gomock.Any()).DoAndReturn(func(pctx *provision.Context, sb sandboxer.Sandbox) (sandboxer.Sandbox, error) {
		return sb, nil
	}).Times(2)

	runtimes := map[string]*config.Runtime{
		"node":   {},
		"python": {},
	}

	p := provision.New[*testRequest](
		runtimes,
		s.op,
	)

	innerLauncherCalled := false
	innerLauncher := func(ctx context.Context, req *testRequest) (sandboxer.Sandbox, error) {
		innerLauncherCalled = true
		s.Require().Equal([]string{"node", "python"}, req.Items) // sorted order
		return s.baseSb, nil
	}

	req := &testRequest{}
	decorated := p.Launch("/opt/runtimes", innerLauncher)
	sb, err := decorated(s.T().Context(), req)
	s.Require().NoError(err)
	s.Require().True(innerLauncherCalled)
	s.Require().NotNil(sb)

	// Verify cleanup was attached to sandbox
	s.baseSb.EXPECT().Terminate(gomock.Any()).Return(nil)
	s.Require().NoError(sb.Terminate(s.T().Context()))
	s.Require().True(cleanedUp.Load())
}

func (s *ProvisionerTestSuite) TestPrepareFailureRollback() {
	var cleanup1Called atomic.Bool

	s.op.EXPECT().Name().Return("fail-op").AnyTimes()
	s.op.EXPECT().Prepare(gomock.Any()).DoAndReturn(func(pctx *provision.Context) (sandboxer.Cleanup, error) {
		if pctx.RuntimeName == "node" {
			return func(context.Context) error {
				cleanup1Called.Store(true)
				return nil
			}, nil
		}
		return nil, errors.New("pull failed")
	}).Times(2)

	runtimes := map[string]*config.Runtime{
		"node":   {},
		"python": {},
	}

	p := provision.New[*testRequest](runtimes, s.op)
	innerLauncher := func(ctx context.Context, req *testRequest) (sandboxer.Sandbox, error) {
		s.T().Fatal("inner launcher should not be called on prepare failure")
		return nil, nil
	}

	_, err := p.Launch("/opt/runtimes", innerLauncher)(s.T().Context(), &testRequest{})
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "pull failed")
	s.Require().True(cleanup1Called.Load(), "rollback should execute cleanup for succeeded runtimes")
}

func (s *ProvisionerTestSuite) TestInnerLauncherFailureRollback() {
	var cleanupCalled atomic.Bool

	s.op.EXPECT().Prepare(gomock.Any()).DoAndReturn(func(pctx *provision.Context) (sandboxer.Cleanup, error) {
		return func(context.Context) error {
			cleanupCalled.Store(true)
			return nil
		}, nil
	}).Times(1)
	s.op.EXPECT().PreLaunch(gomock.Any(), gomock.Any()).DoAndReturn(func(pctx *provision.Context, req *testRequest) (*testRequest, error) {
		return req, nil
	}).Times(1)

	runtimes := map[string]*config.Runtime{"node": {}}
	p := provision.New[*testRequest](runtimes, s.op)

	innerLauncher := func(ctx context.Context, req *testRequest) (sandboxer.Sandbox, error) {
		return nil, errors.New("container run error")
	}

	_, err := p.Launch("/opt/runtimes", innerLauncher)(s.T().Context(), &testRequest{})
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "container run error")
	s.Require().True(cleanupCalled.Load(), "rollback should execute cleanups when inner launcher fails")
}

func (s *ProvisionerTestSuite) TestPreLaunchFailureRollback() {
	var cleanupCalled atomic.Bool

	s.op.EXPECT().Name().Return("fail-op").AnyTimes()
	s.op.EXPECT().Prepare(gomock.Any()).DoAndReturn(func(pctx *provision.Context) (sandboxer.Cleanup, error) {
		return func(context.Context) error {
			cleanupCalled.Store(true)
			return nil
		}, nil
	}).Times(1)
	s.op.EXPECT().PreLaunch(gomock.Any(), gomock.Any()).Return(nil, errors.New("prelaunch mutation failed")).Times(1)

	runtimes := map[string]*config.Runtime{"node": {}}
	p := provision.New[*testRequest](runtimes, s.op)

	innerLauncher := func(ctx context.Context, req *testRequest) (sandboxer.Sandbox, error) {
		s.T().Fatal("inner launcher should not be called on prelaunch failure")
		return nil, nil
	}

	_, err := p.Launch("/opt/runtimes", innerLauncher)(s.T().Context(), &testRequest{})
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "prelaunch mutation failed")
	s.Require().True(cleanupCalled.Load(), "rollback should execute cleanups when pre-launch fails")
}

func (s *ProvisionerTestSuite) TestPostLaunchFailureRollback() {
	var cleanupCalled atomic.Bool

	s.op.EXPECT().Name().Return("fail-op").AnyTimes()
	s.op.EXPECT().Prepare(gomock.Any()).DoAndReturn(func(pctx *provision.Context) (sandboxer.Cleanup, error) {
		return func(context.Context) error {
			cleanupCalled.Store(true)
			return nil
		}, nil
	}).Times(1)
	s.op.EXPECT().PreLaunch(gomock.Any(), gomock.Any()).DoAndReturn(func(pctx *provision.Context, req *testRequest) (*testRequest, error) {
		return req, nil
	}).Times(1)
	s.op.EXPECT().PostLaunch(gomock.Any(), gomock.Any()).Return(nil, errors.New("postlaunch failed")).Times(1)

	runtimes := map[string]*config.Runtime{"node": {}}
	p := provision.New[*testRequest](runtimes, s.op)

	s.baseSb.EXPECT().Terminate(gomock.Any()).Return(nil)

	innerLauncher := func(ctx context.Context, req *testRequest) (sandboxer.Sandbox, error) {
		return s.baseSb, nil
	}

	_, err := p.Launch("/opt/runtimes", innerLauncher)(s.T().Context(), &testRequest{})
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "postlaunch failed")
	s.Require().True(cleanupCalled.Load(), "rollback should execute cleanups when post-launch fails")
}

func (s *ProvisionerTestSuite) TestPassthrough() {
	innerLauncher := func(ctx context.Context, req *testRequest) (sandboxer.Sandbox, error) {
		return s.baseSb, nil
	}

	// 1. Nil provisioner
	var pNil *provision.Provisioner[*testRequest]
	sb, err := pNil.Launch("/opt/runtimes", innerLauncher)(s.T().Context(), &testRequest{})
	s.Require().NoError(err)
	s.Require().Equal(s.baseSb, sb)

	// 2. Empty runtimes
	pEmptyRuntimes := provision.New[*testRequest](nil)
	sb, err = pEmptyRuntimes.Launch("/opt/runtimes", innerLauncher)(s.T().Context(), &testRequest{})
	s.Require().NoError(err)
	s.Require().Equal(s.baseSb, sb)

	// 3. Empty ops
	pEmptyOps := provision.New[*testRequest](map[string]*config.Runtime{"node": {}})
	sb, err = pEmptyOps.Launch("/opt/runtimes", innerLauncher)(s.T().Context(), &testRequest{})
	s.Require().NoError(err)
	s.Require().Equal(s.baseSb, sb)
}

func (s *ProvisionerTestSuite) TestLIFOCleanupOrder() {
	var mu sync.Mutex
	var order []int

	op1 := mock_provision.NewMockOperation[*testRequest](s.ctrl)
	op1.EXPECT().Prepare(gomock.Any()).DoAndReturn(func(pctx *provision.Context) (sandboxer.Cleanup, error) {
		return func(ctx context.Context) error {
			mu.Lock()
			order = append(order, 1)
			mu.Unlock()
			return nil
		}, nil
	}).Times(1)
	op1.EXPECT().PreLaunch(gomock.Any(), gomock.Any()).DoAndReturn(func(pctx *provision.Context, req *testRequest) (*testRequest, error) {
		return req, nil
	}).Times(1)
	op1.EXPECT().PostLaunch(gomock.Any(), gomock.Any()).DoAndReturn(func(pctx *provision.Context, sb sandboxer.Sandbox) (sandboxer.Sandbox, error) {
		return sb, nil
	}).Times(1)

	op2 := mock_provision.NewMockOperation[*testRequest](s.ctrl)
	op2.EXPECT().Prepare(gomock.Any()).DoAndReturn(func(pctx *provision.Context) (sandboxer.Cleanup, error) {
		return func(ctx context.Context) error {
			mu.Lock()
			order = append(order, 2)
			mu.Unlock()
			return nil
		}, nil
	}).Times(1)
	op2.EXPECT().PreLaunch(gomock.Any(), gomock.Any()).DoAndReturn(func(pctx *provision.Context, req *testRequest) (*testRequest, error) {
		return req, nil
	}).Times(1)
	op2.EXPECT().PostLaunch(gomock.Any(), gomock.Any()).DoAndReturn(func(pctx *provision.Context, sb sandboxer.Sandbox) (sandboxer.Sandbox, error) {
		return sb, nil
	}).Times(1)

	p := provision.New[*testRequest](
		map[string]*config.Runtime{"node": {}},
		op1,
		op2,
	)

	s.baseSb.EXPECT().Terminate(gomock.Any()).Return(nil)
	innerLauncher := func(ctx context.Context, req *testRequest) (sandboxer.Sandbox, error) {
		return s.baseSb, nil
	}

	sb, err := p.Launch("/opt/runtimes", innerLauncher)(s.T().Context(), &testRequest{})
	s.Require().NoError(err)

	s.Require().NoError(sb.Terminate(s.T().Context()))

	mu.Lock()
	defer mu.Unlock()
	s.Require().Equal([]int{2, 1}, order, "cleanups should execute in LIFO order (reverse of registration)")
}

func (s *ProvisionerTestSuite) TestPrepareCancellation() {
	canceled := make(chan struct{})

	s.op.EXPECT().Name().Return("cancel-test-op").AnyTimes()
	s.op.EXPECT().Prepare(gomock.Any()).DoAndReturn(func(pctx *provision.Context) (sandboxer.Cleanup, error) {
		if pctx.RuntimeName == "fast-fail" {
			return nil, errors.New("boom")
		}
		// sibling runtime should see cancellation
		select {
		case <-pctx.Done():
			close(canceled)
		case <-time.After(2 * time.Second):
			s.T().Error("timed out waiting for sibling context cancellation")
		}
		return nil, nil
	}).Times(2)

	p := provision.New[*testRequest](
		map[string]*config.Runtime{
			"fast-fail": {},
			"slow":      {},
		},
		s.op,
	)

	_, err := p.Launch("/opt/runtimes", func(ctx context.Context, req *testRequest) (sandboxer.Sandbox, error) {
		return nil, nil
	})(s.T().Context(), &testRequest{})

	s.Require().Error(err)
	select {
	case <-canceled:
		// success: sibling was canceled
	default:
		s.T().Fatal("expected sibling runtime to receive context cancellation")
	}
}
