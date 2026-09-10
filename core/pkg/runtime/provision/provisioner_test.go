/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package provision_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"drassi.run/core/config"
	mock_sandboxer "drassi.run/core/mock/sandboxer"
	"drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/sandboxer"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type testRequest struct {
	Items []string
}

type mockOp struct {
	provision.Noop[*testRequest]
	name         string
	prepareFn    func(pctx *provision.Context) (sandboxer.Cleanup, error)
	preLaunchFn  func(pctx *provision.Context, req *testRequest) (*testRequest, error)
	postLaunchFn func(pctx *provision.Context, sb sandboxer.Sandbox) (sandboxer.Sandbox, error)
}

func (m *mockOp) Name() string { return m.name }
func (m *mockOp) Prepare(pctx *provision.Context) (sandboxer.Cleanup, error) {
	if m.prepareFn != nil {
		return m.prepareFn(pctx)
	}
	return nil, nil
}
func (m *mockOp) PreLaunch(pctx *provision.Context, req *testRequest) (*testRequest, error) {
	if m.preLaunchFn != nil {
		return m.preLaunchFn(pctx, req)
	}
	return req, nil
}
func (m *mockOp) PostLaunch(pctx *provision.Context, sb sandboxer.Sandbox) (sandboxer.Sandbox, error) {
	if m.postLaunchFn != nil {
		return m.postLaunchFn(pctx, sb)
	}
	return sb, nil
}

func TestProvisionerSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	baseSb := mock_sandboxer.NewMockSandbox(ctrl)

	var cleanedUp atomic.Bool
	cleanupFn := func(context.Context) error {
		cleanedUp.Store(true)
		return nil
	}

	op := &mockOp{
		name: "test-op",
		prepareFn: func(pctx *provision.Context) (sandboxer.Cleanup, error) {
			return cleanupFn, nil
		},
		preLaunchFn: func(pctx *provision.Context, req *testRequest) (*testRequest, error) {
			req.Items = append(req.Items, pctx.RuntimeName)
			return req, nil
		},
	}

	runtimes := map[string]*config.Runtime{
		"node":   {},
		"python": {},
	}

	p := provision.New[*testRequest](
		runtimes,
		func(name string) string { return "/opt/" + name },
		op,
	)

	innerLauncherCalled := false
	innerLauncher := func(ctx context.Context, req *testRequest) (sandboxer.Sandbox, error) {
		innerLauncherCalled = true
		require.Equal(t, []string{"node", "python"}, req.Items) // sorted order
		return baseSb, nil
	}

	req := &testRequest{}
	decorated := p.Launch(t.Context(), innerLauncher)
	sb, err := decorated(t.Context(), req)
	require.NoError(t, err)
	require.True(t, innerLauncherCalled)
	require.NotNil(t, sb)

	// Verify cleanup was attached to sandbox
	baseSb.EXPECT().Terminate(gomock.Any()).Return(nil)
	require.NoError(t, sb.Terminate(t.Context()))
	require.True(t, cleanedUp.Load())
}

func TestProvisionerPrepareFailureRollback(t *testing.T) {
	var cleanup1Called atomic.Bool
	op := &mockOp{
		name: "fail-op",
		prepareFn: func(pctx *provision.Context) (sandboxer.Cleanup, error) {
			if pctx.RuntimeName == "node" {
				return func(context.Context) error {
					cleanup1Called.Store(true)
					return nil
				}, nil
			}
			return nil, errors.New("pull failed")
		},
	}

	runtimes := map[string]*config.Runtime{
		"node":   {},
		"python": {},
	}

	p := provision.New[*testRequest](runtimes, nil, op)
	innerLauncher := func(ctx context.Context, req *testRequest) (sandboxer.Sandbox, error) {
		t.Fatal("inner launcher should not be called on prepare failure")
		return nil, nil
	}

	_, err := p.Launch(t.Context(), innerLauncher)(t.Context(), &testRequest{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "pull failed")
	require.True(t, cleanup1Called.Load(), "rollback should execute cleanup for succeeded runtimes")
}

func TestProvisionerInnerLauncherFailureRollback(t *testing.T) {
	var cleanupCalled atomic.Bool
	op := &mockOp{
		name: "test-op",
		prepareFn: func(pctx *provision.Context) (sandboxer.Cleanup, error) {
			return func(context.Context) error {
				cleanupCalled.Store(true)
				return nil
			}, nil
		},
	}

	runtimes := map[string]*config.Runtime{"node": {}}
	p := provision.New[*testRequest](runtimes, nil, op)

	innerLauncher := func(ctx context.Context, req *testRequest) (sandboxer.Sandbox, error) {
		return nil, errors.New("container run error")
	}

	_, err := p.Launch(t.Context(), innerLauncher)(t.Context(), &testRequest{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "container run error")
	require.True(t, cleanupCalled.Load(), "rollback should execute cleanups when inner launcher fails")
}

func TestProvisionerPreLaunchFailureRollback(t *testing.T) {
	var cleanupCalled atomic.Bool
	op := &mockOp{
		name: "test-op",
		prepareFn: func(pctx *provision.Context) (sandboxer.Cleanup, error) {
			return func(context.Context) error {
				cleanupCalled.Store(true)
				return nil
			}, nil
		},
		preLaunchFn: func(pctx *provision.Context, req *testRequest) (*testRequest, error) {
			return nil, errors.New("prelaunch mutation failed")
		},
	}

	runtimes := map[string]*config.Runtime{"node": {}}
	p := provision.New[*testRequest](runtimes, nil, op)

	innerLauncher := func(ctx context.Context, req *testRequest) (sandboxer.Sandbox, error) {
		t.Fatal("inner launcher should not be called on prelaunch failure")
		return nil, nil
	}

	_, err := p.Launch(t.Context(), innerLauncher)(t.Context(), &testRequest{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "prelaunch mutation failed")
	require.True(t, cleanupCalled.Load(), "rollback should execute cleanups when pre-launch fails")
}

func TestProvisionerPostLaunchFailureRollback(t *testing.T) {
	ctrl := gomock.NewController(t)
	baseSb := mock_sandboxer.NewMockSandbox(ctrl)

	var cleanupCalled atomic.Bool
	op := &mockOp{
		name: "test-op",
		prepareFn: func(pctx *provision.Context) (sandboxer.Cleanup, error) {
			return func(context.Context) error {
				cleanupCalled.Store(true)
				return nil
			}, nil
		},
		postLaunchFn: func(pctx *provision.Context, sb sandboxer.Sandbox) (sandboxer.Sandbox, error) {
			return nil, errors.New("postlaunch failed")
		},
	}

	runtimes := map[string]*config.Runtime{"node": {}}
	p := provision.New[*testRequest](runtimes, nil, op)

	baseSb.EXPECT().Terminate(gomock.Any()).Return(nil)

	innerLauncher := func(ctx context.Context, req *testRequest) (sandboxer.Sandbox, error) {
		return baseSb, nil
	}

	_, err := p.Launch(t.Context(), innerLauncher)(t.Context(), &testRequest{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "postlaunch failed")
	require.True(t, cleanupCalled.Load(), "rollback should execute cleanups when post-launch fails")
}

func TestProvisionerPassthrough(t *testing.T) {
	ctrl := gomock.NewController(t)
	baseSb := mock_sandboxer.NewMockSandbox(ctrl)

	innerLauncher := func(ctx context.Context, req *testRequest) (sandboxer.Sandbox, error) {
		return baseSb, nil
	}

	// 1. Nil provisioner
	var pNil *provision.Provisioner[*testRequest]
	sb, err := pNil.Launch(t.Context(), innerLauncher)(t.Context(), &testRequest{})
	require.NoError(t, err)
	require.Equal(t, baseSb, sb)

	// 2. Empty runtimes
	pEmptyRuntimes := provision.New[*testRequest](nil, nil)
	sb, err = pEmptyRuntimes.Launch(t.Context(), innerLauncher)(t.Context(), &testRequest{})
	require.NoError(t, err)
	require.Equal(t, baseSb, sb)

	// 3. Empty ops
	pEmptyOps := provision.New[*testRequest](map[string]*config.Runtime{"node": {}}, nil)
	sb, err = pEmptyOps.Launch(t.Context(), innerLauncher)(t.Context(), &testRequest{})
	require.NoError(t, err)
	require.Equal(t, baseSb, sb)
}
