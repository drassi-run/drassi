/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package firecracker

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/stream"
	"drassi.run/core/util/tar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVMLaunchCopyExec(t *testing.T) {
	requireFirecracker(t)
	kernel, rootfs := guestImages(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cfg := DefaultConfig()
	cfg.Bin = "firecracker"
	cfg.kernel = kernel
	cfg.RootfsPath = ""
	cfg.rootfs = rootfs
	cfg.RootDir = t.TempDir()
	cfg.VcpuCount = 1
	cfg.MemSizeMiB = 256
	cfg.AgentWait = 60
	cfg.KernelArgs = "console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda rw init=/sbin/init"

	engine, err := New(cfg)
	require.NoError(t, err)
	defer engine.Close()

	resp, err := engine.Launch(ctx, &sandboxer.LaunchRequest{Uid: "vm-copy-exec"})
	if err != nil {
		dumpVMLog(t, filepath.Join(cfg.RootDir, "vm-copy-exec"))
		require.NoError(t, err)
	}
	sb := resp.Sandbox
	t.Cleanup(func() {
		_ = sb.Terminate(context.Background())
	})

	t.Run("stat-layout", func(t *testing.T) {
		info, err := sb.Stat(ctx, sb.Layout().Workspace)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("copy-in-out", func(t *testing.T) {
		r, err := xtar.ContentReader(map[string]string{"hello.txt": "from-host"})
		require.NoError(t, err)
		require.NoError(t, sb.CopyIn(ctx, r, sb.Layout().Temp))

		guestPath := path.Join(sb.Layout().Temp, "hello.txt")
		info, err := sb.Stat(ctx, guestPath)
		require.NoError(t, err)
		assert.Equal(t, "hello.txt", info.Name())
		assert.Equal(t, int64(len("from-host")), info.Size())

		out, err := sb.CopyOut(ctx, guestPath)
		require.NoError(t, err)
		defer out.Close()

		var content string
		err = xtar.Untar(ctx, out, func(hdr *tar.Header, r io.Reader) error {
			if !xtar.IsRegular(hdr) {
				return nil
			}
			b, err := io.ReadAll(r)
			if err != nil {
				return err
			}
			content = string(b)
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, "from-host", content)
	})

	t.Run("copy-out-missing", func(t *testing.T) {
		_, err := sb.CopyOut(ctx, "/no/such/file")
		require.Error(t, err)
		assert.True(t, errors.Is(err, fs.ErrNotExist))
	})

	t.Run("execute", func(t *testing.T) {
		var stdout bytes.Buffer
		err := sb.Execute(ctx, []string{"echo", "-n", "hi-vm"}, nil, nil, "", &stream.Streams{
			Out: &stdout,
		})
		require.NoError(t, err)
		assert.Equal(t, "hi-vm", stdout.String())
	})

	t.Run("execute-failure", func(t *testing.T) {
		err := sb.Execute(ctx, []string{"false"}, nil, nil, "", &stream.Streams{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exitcode")
	})
}

func TestVMHybridVsockAgent(t *testing.T) {
	requireFirecracker(t)
	kernel, rootfs := guestImages(t)

	dir := t.TempDir()
	cfg := DefaultConfig()
	cfg.kernel = kernel
	cfg.RootfsPath = ""
	cfg.rootfs = rootfs
	cfg.VcpuCount = 1
	cfg.MemSizeMiB = 256
	cfg.AgentWait = 60
	cfg.KernelArgs = "console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda rw init=/sbin/init"

	vmDir := filepath.Join(dir, "guest")
	require.NoError(t, os.MkdirAll(vmDir, 0o755))

	machine, err := startVM(cfg, vmDir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = machine.stop(context.Background()) })

	cl := &client{
		dial: func(ctx context.Context) (net.Conn, error) {
			return DialHybridVsock(ctx, machine.vsockPath(), cfg.AgentPort)
		},
	}

	e := &engine{Config: *cfg}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err = e.waitAgent(ctx, cl, machine); err != nil {
		dumpVMLog(t, vmDir)
		require.NoError(t, err)
	}

	pathEnv, err := cl.Info(ctx)
	require.NoError(t, err)
	assert.True(t, strings.Contains(pathEnv, "/bin") || pathEnv == "")

	r, err := xtar.ContentReader(map[string]string{"ping.txt": "pong"})
	require.NoError(t, err)
	require.NoError(t, cl.CopyIn(ctx, r, "/tmp"))

	info, err := cl.Stat(ctx, "/tmp/ping.txt")
	require.NoError(t, err)
	assert.Equal(t, int64(4), info.Size())
}
