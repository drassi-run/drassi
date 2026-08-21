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
	"path/filepath"
	"testing"

	"drassi.run/core/pkg/stream"
	"drassi.run/core/util/tar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func startTestAgent(t *testing.T) *client {
	t.Helper()

	dir := t.TempDir()
	sock := filepath.Join(dir, "agent.sock")
	ln, err := net.Listen("unix", sock)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	errCh := make(chan error, 1)
	go func() {
		errCh <- Serve(ctx, ln)
	}()
	t.Cleanup(func() {
		cancel()
		_ = ln.Close()
		<-errCh
	})

	return &client{
		dial: func(ctx context.Context) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", sock)
		},
	}
}

func TestCopyInOutStat(t *testing.T) {
	cl := startTestAgent(t)
	ctx := context.Background()
	dir := t.TempDir()

	r, err := xtar.ContentReader(map[string]string{
		"hello.txt": "world",
	})
	require.NoError(t, err)
	require.NoError(t, cl.CopyIn(ctx, r, dir))

	info, err := cl.Stat(ctx, filepath.Join(dir, "hello.txt"))
	require.NoError(t, err)
	assert.Equal(t, "hello.txt", info.Name())
	assert.False(t, info.IsDir())
	assert.Equal(t, int64(5), info.Size())

	out, err := cl.CopyOut(ctx, filepath.Join(dir, "hello.txt"))
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
	assert.Equal(t, "world", content)
}

func TestCopyOutMissing(t *testing.T) {
	cl := startTestAgent(t)
	_, err := cl.CopyOut(context.Background(), filepath.Join(t.TempDir(), "missing"))
	require.Error(t, err)
	assert.True(t, errors.Is(err, fs.ErrNotExist))
}

func TestStatMissing(t *testing.T) {
	cl := startTestAgent(t)
	_, err := cl.Stat(context.Background(), filepath.Join(t.TempDir(), "missing"))
	require.Error(t, err)
	assert.True(t, errors.Is(err, fs.ErrNotExist))
}

func TestExecute(t *testing.T) {
	cl := startTestAgent(t)
	ctx := context.Background()

	var stdout bytes.Buffer
	err := cl.Execute(ctx, []string{"echo", "-n", "hi"}, nil, nil, "", &stream.Streams{
		Out: &stdout,
	})
	require.NoError(t, err)
	assert.Equal(t, "hi", stdout.String())
}

func TestExecuteFailure(t *testing.T) {
	cl := startTestAgent(t)
	err := cl.Execute(context.Background(), []string{"false"}, nil, nil, "", &stream.Streams{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exitcode")
}

func TestInfo(t *testing.T) {
	cl := startTestAgent(t)
	path, err := cl.Info(context.Background())
	require.NoError(t, err)
	assert.Equal(t, os.Getenv("PATH"), path)
}

func TestHybridVsockDial(t *testing.T) {
	dir := t.TempDir()
	sock := filepath.Join(dir, "vsock.sock")
	ln, err := net.Listen("unix", sock)
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		line, err := readLine(conn)
		if err != nil {
			return
		}
		if line != "CONNECT 1024" {
			return
		}
		_, _ = conn.Write([]byte("OK 1073741824\n"))
		buf := make([]byte, 4)
		_, _ = io.ReadFull(conn, buf)
		_, _ = conn.Write(buf)
	}()

	ctx := context.Background()
	conn, err := DialHybridVsock(ctx, sock, 1024)
	require.NoError(t, err)
	defer conn.Close()

	_, err = conn.Write([]byte("ping"))
	require.NoError(t, err)
	buf := make([]byte, 4)
	_, err = io.ReadFull(conn, buf)
	require.NoError(t, err)
	assert.Equal(t, "ping", string(buf))
	<-done
}
