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

type mockOp struct {
	name            string
	preCalled       bool
	postCalled      bool
	preErr          error
	postErr         error
	receivedPreCtx  *provision.Context
	receivedPostCtx *provision.Context
	receivedSb      sandboxer.Sandbox
}

func (m *mockOp) Name() string { return m.name }
func (m *mockOp) PreLaunch(pctx *provision.Context) error {
	m.preCalled = true
	m.receivedPreCtx = pctx
	return m.preErr
}
func (m *mockOp) PostLaunch(pctx *provision.Context, sb sandboxer.Sandbox) error {
	m.postCalled = true
	m.receivedPostCtx = pctx
	m.receivedSb = sb
	return m.postErr
}

func TestPipelineExecution(t *testing.T) {
	op1 := &mockOp{name: "op1"}
	op2 := &mockOp{name: "op2"}

	p := provision.NewPipeline(op1, op2)
	pctx := provision.NewContext(context.Background(), "node", &config.Runtime{}, "/target")

	err := p.PreLaunch(pctx)
	require.NoError(t, err)
	require.True(t, op1.preCalled)
	require.True(t, op2.preCalled)
	require.Equal(t, pctx, op1.receivedPreCtx)
	require.Equal(t, pctx, op2.receivedPreCtx)

	err = p.PostLaunch(pctx, nil)
	require.NoError(t, err)
	require.True(t, op1.postCalled)
	require.True(t, op2.postCalled)
	require.Equal(t, pctx, op1.receivedPostCtx)
	require.Equal(t, pctx, op2.receivedPostCtx)
}

func TestPipelinePreLaunchError(t *testing.T) {
	op1 := &mockOp{name: "op1", preErr: errors.New("pre fail")}
	op2 := &mockOp{name: "op2"}

	p := provision.NewPipeline(op1, op2)
	pctx := provision.NewContext(context.Background(), "node", &config.Runtime{}, "/target")

	err := p.PreLaunch(pctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), `operation "op1" pre-launch failed for runtime "node": pre fail`)
	require.False(t, op2.preCalled)
}

func TestPipelinePostLaunchError(t *testing.T) {
	op1 := &mockOp{name: "op1"}
	op2 := &mockOp{name: "op2", postErr: errors.New("post fail")}

	p := provision.NewPipeline(op1, op2)
	pctx := provision.NewContext(context.Background(), "python", &config.Runtime{}, "/target")

	err := p.PostLaunch(pctx, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), `operation "op2" post-launch failed for runtime "python": post fail`)
	require.False(t, op1.postCalled)
}

func TestPipelineEmpty(t *testing.T) {
	p := provision.NewPipeline()
	pctx := provision.NewContext(context.Background(), "node", &config.Runtime{}, "/target")

	require.NoError(t, p.PreLaunch(pctx))
	require.NoError(t, p.PostLaunch(pctx, nil))
}
