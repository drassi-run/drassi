/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package provision_test

import (
	"context"
	"errors"
	"testing"

	"drassi.run/core/config"
	"drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/sandboxer"
	"github.com/stretchr/testify/require"
)

func TestPipelineExecution(t *testing.T) {
	var (
		op1PreCalled, op2PreCalled   bool
		op1PostCalled, op2PostCalled bool
		op1PreCtx, op2PreCtx         *provision.Context
		op1PostCtx, op2PostCtx       *provision.Context
	)

	op1 := &provision.OpFunc{
		PreFunc: func(pctx *provision.Context) error {
			op1PreCalled = true
			op1PreCtx = pctx
			return nil
		},
		PostFunc: func(pctx *provision.Context, _ sandboxer.Sandbox) error {
			op1PostCalled = true
			op1PostCtx = pctx
			return nil
		},
	}
	op2 := &provision.OpFunc{
		PreFunc: func(pctx *provision.Context) error {
			op2PreCalled = true
			op2PreCtx = pctx
			return nil
		},
		PostFunc: func(pctx *provision.Context, _ sandboxer.Sandbox) error {
			op2PostCalled = true
			op2PostCtx = pctx
			return nil
		},
	}

	p := provision.NewPipeline(op1, op2)
	pctx := provision.NewContext(context.Background(), "node", &config.Runtime{}, "/target")

	err := p.PreLaunch(pctx)
	require.NoError(t, err)
	require.True(t, op1PreCalled)
	require.True(t, op2PreCalled)
	require.Equal(t, pctx, op1PreCtx)
	require.Equal(t, pctx, op2PreCtx)

	err = p.PostLaunch(pctx, nil)
	require.NoError(t, err)
	require.True(t, op1PostCalled)
	require.True(t, op2PostCalled)
	require.Equal(t, pctx, op1PostCtx)
	require.Equal(t, pctx, op2PostCtx)
}

func TestPipelinePreLaunchError(t *testing.T) {
	op2PreCalled := false
	op1 := &provision.OpFunc{
		PreFunc: func(_ *provision.Context) error {
			return errors.New("pre fail")
		},
	}
	op2 := &provision.OpFunc{
		PreFunc: func(_ *provision.Context) error {
			op2PreCalled = true
			return nil
		},
	}

	p := provision.NewPipeline(op1, op2)
	pctx := provision.NewContext(context.Background(), "node", &config.Runtime{}, "/target")

	err := p.PreLaunch(pctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), `operation "func" pre-launch failed for runtime "node": pre fail`)
	require.False(t, op2PreCalled)
}

func TestPipelinePostLaunchError(t *testing.T) {
	op1PostCalled := false
	op1 := &provision.OpFunc{
		PostFunc: func(_ *provision.Context, _ sandboxer.Sandbox) error {
			op1PostCalled = true
			return nil
		},
	}
	op2 := &provision.OpFunc{
		PostFunc: func(_ *provision.Context, _ sandboxer.Sandbox) error {
			return errors.New("post fail")
		},
	}

	p := provision.NewPipeline(op1, op2)
	pctx := provision.NewContext(context.Background(), "python", &config.Runtime{}, "/target")

	err := p.PostLaunch(pctx, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), `operation "func" post-launch failed for runtime "python": post fail`)
	require.False(t, op1PostCalled)
}

func TestPipelineEmpty(t *testing.T) {
	p := provision.NewPipeline()
	pctx := provision.NewContext(context.Background(), "node", &config.Runtime{}, "/target")

	require.NoError(t, p.PreLaunch(pctx))
	require.NoError(t, p.PostLaunch(pctx, nil))
}
