/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package firecracker

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"drassi.run/core/util/fs"
	"drassi.run/core/util/path"
	"github.com/go-git/go-billy/v5/osfs"
	"golang.org/x/sync/errgroup"
)

// Serve runs the guest agent on ln. The Firecracker host dials this listener
// through hybrid vsock; tests can serve on a Unix socket instead.
func Serve(ctx context.Context, ln net.Listener) error {
	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		go handle(ctx, conn)
	}
}

func handle(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	s := newSession(conn)
	var req request
	if err := s.readJSON(&req); err != nil {
		return
	}

	switch req.Op {
	case opCopyIn:
		handleCopyIn(ctx, s, req)
	case opCopyOut:
		handleCopyOut(ctx, s, req)
	case opStat:
		handleStat(s, req)
	case opExec:
		handleExec(ctx, s, req)
	case opInfo:
		_ = s.writeJSON(response{OK: true, EnvPath: os.Getenv("PATH")})
	case opNet:
		handleNet(s, req)
	default:
		_ = s.writeJSON(response{Error: fmt.Sprintf("unknown op %q", req.Op)})
	}
}

func handleNet(s *session, req request) {
	if err := applyGuestNet(req); err != nil {
		_ = s.writeJSON(errResponse(err))
		return
	}
	_ = s.writeJSON(okResponse())
}

func applyGuestNet(req request) error {
	iface := req.Iface
	if iface == "" {
		iface = "eth0"
	}
	if req.Addr == "" {
		return fmt.Errorf("guest addr is required")
	}
	if err := ipCmd("link", "set", iface, "up"); err != nil {
		return err
	}
	if err := ipCmd("addr", "replace", req.Addr, "dev", iface); err != nil {
		return err
	}
	if req.Gateway != "" {
		if err := ipCmd("route", "replace", "default", "via", req.Gateway); err != nil {
			return err
		}
	}
	if len(req.DNS) == 0 {
		return nil
	}
	var b strings.Builder
	for _, d := range req.DNS {
		if d == "" {
			continue
		}
		fmt.Fprintf(&b, "nameserver %s\n", d)
	}
	if b.Len() == 0 {
		return nil
	}
	return os.WriteFile("/etc/resolv.conf", []byte(b.String()), 0o644)
}

func ipCmd(args ...string) error {
	cmd := exec.Command("ip", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ip %s: %w: %s", strings.Join(args, " "), err, bytes.TrimSpace(out))
	}
	return nil
}

func handleCopyIn(ctx context.Context, s *session, req request) {
	fsys := osfs.New("/")
	err := xfs.Write(ctx, fsys, s.r, req.Path)
	if err != nil {
		_ = s.writeJSON(errResponse(err))
		return
	}
	_ = s.writeJSON(okResponse())
}

func handleCopyOut(ctx context.Context, s *session, req request) {
	if _, err := os.Stat(req.Path); err != nil {
		_ = s.writeJSON(errResponse(err))
		return
	}
	if err := s.writeJSON(okResponse()); err != nil {
		return
	}
	fsys := osfs.New("/")
	_, _ = io.Copy(s.conn, xfs.Read(ctx, fsys, req.Path))
}

func handleStat(s *session, req request) {
	info, err := os.Stat(req.Path)
	if err != nil {
		_ = s.writeJSON(errResponse(err))
		return
	}
	_ = s.writeJSON(response{
		OK:      true,
		Name:    info.Name(),
		Size:    info.Size(),
		Mode:    uint32(info.Mode()),
		ModTime: info.ModTime().UnixNano(),
	})
}

func handleExec(ctx context.Context, s *session, req request) {
	if len(req.Cmd) == 0 {
		_ = s.writeJSON(response{Error: "empty command"})
		return
	}

	c := exec.CommandContext(ctx, req.Cmd[0], req.Cmd[1:]...)
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	c.Cancel = func() error {
		return syscall.Kill(-c.Process.Pid, syscall.SIGKILL)
	}

	c.Env = os.Environ()
	for k, v := range req.Env {
		c.Env = append(c.Env, k+"="+v)
	}
	if p := os.Getenv("PATH"); p != "" {
		req.PathEnv = append(req.PathEnv, p)
	}
	if len(req.PathEnv) > 0 {
		c.Env = append(c.Env, "PATH="+strings.Join(req.PathEnv, string(os.PathListSeparator)))
	}
	if req.Workdir != "" {
		c.Dir = xpath.Abs(req.Workdir, "/")
	}

	stdin, err := c.StdinPipe()
	if err != nil {
		_ = s.writeJSON(errResponse(err))
		return
	}
	stdout, err := c.StdoutPipe()
	if err != nil {
		_ = s.writeJSON(errResponse(err))
		return
	}
	stderr, err := c.StderrPipe()
	if err != nil {
		_ = s.writeJSON(errResponse(err))
		return
	}
	if err = c.Start(); err != nil {
		_ = s.writeJSON(errResponse(err))
		return
	}
	if err = s.writeJSON(okResponse()); err != nil {
		_ = c.Process.Kill()
		return
	}

	go func() {
		defer stdin.Close()
		for {
			stream, payload, err := s.readFrame()
			if err != nil {
				return
			}
			if stream != streamStdin {
				continue
			}
			if len(payload) == 0 {
				return
			}
			if _, err = stdin.Write(payload); err != nil {
				return
			}
		}
	}()

	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error { return copyToFrames(s, streamStdout, stdout) })
	g.Go(func() error { return copyToFrames(s, streamStderr, stderr) })
	_ = g.Wait()

	exitCode := 0
	if err = c.Wait(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			exitCode = 1
		}
	}

	var code [4]byte
	binary.BigEndian.PutUint32(code[:], uint32(exitCode))
	_ = s.writeFrame(streamExit, code[:])
}

func copyToFrames(s *session, stream byte, r io.Reader) error {
	buf := make([]byte, 32*1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			if werr := s.writeFrame(stream, buf[:n]); werr != nil {
				return werr
			}
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}
