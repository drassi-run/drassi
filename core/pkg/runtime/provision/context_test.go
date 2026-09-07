package provision_test

import (
	"context"
	"testing"
	"time"

	"drassi.run/core/config"
	"drassi.run/core/pkg/runtime/provision"
	"github.com/stretchr/testify/require"
)

func TestContextGenericState(t *testing.T) {
	ctx := context.Background()
	rtCfg := &config.Runtime{Image: "drassi/node:24"}
	pctx := provision.NewContext(ctx, "node", rtCfg, "/opt/drassi/runtimes/node")

	require.Equal(t, "node", pctx.RuntimeName)
	require.Equal(t, "/opt/drassi/runtimes/node", pctx.TargetDir)
	require.Equal(t, rtCfg, pctx.Config)

	// Test typed key string
	pctx.Set(provision.KeyHostMountDir, "/var/lib/drassi/storage/overlay/merged")
	val, ok := pctx.Get(provision.KeyHostMountDir)
	require.True(t, ok)
	require.Equal(t, "/var/lib/drassi/storage/overlay/merged", val)
	require.Equal(t, "/var/lib/drassi/storage/overlay/merged", pctx.MustGet(provision.KeyHostMountDir))

	// Test KeyMountID
	pctx.Set(provision.KeyMountID, "mount-12345")
	mountID, ok := pctx.Get(provision.KeyMountID)
	require.True(t, ok)
	require.Equal(t, "mount-12345", mountID)
	require.Equal(t, "mount-12345", pctx.MustGet(provision.KeyMountID))

	// Test missing key
	const keyMissing = provision.StateKey[int]("missing_key")
	intVal, ok := pctx.Get(keyMissing)
	require.False(t, ok)
	require.Equal(t, 0, intVal)
	require.Panics(t, func() {
		pctx.MustGet(keyMissing)
	})

	// Test type mismatch
	const keyMismatch = provision.StateKey[int]("host_mount_dir")
	mismatchVal, ok := pctx.Get(keyMismatch)
	require.False(t, ok)
	require.Equal(t, 0, mismatchVal)
	require.Panics(t, func() {
		pctx.MustGet(keyMismatch)
	})
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	pctx := provision.NewContext(ctx, "node", nil, "")

	select {
	case <-pctx.Done():
		t.Fatal("context should not be done yet")
	default:
	}

	cancel()

	select {
	case <-pctx.Done():
		require.Equal(t, context.Canceled, pctx.Err())
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for context cancellation")
	}
}
