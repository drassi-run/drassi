/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package firecracker

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"net"

	"drassi.run/core/pkg/stream"
)

type client struct {
	dial func(ctx context.Context) (net.Conn, error)
}

func (c *client) call(ctx context.Context) (*session, error) {
	conn, err := c.dial(ctx)
	if err != nil {
		return nil, err
	}
	return newSession(conn), nil
}

func (c *client) Info(ctx context.Context) (string, error) {
	s, err := c.call(ctx)
	if err != nil {
		return "", err
	}
	defer s.conn.Close()

	if err = s.writeJSON(request{Op: opInfo}); err != nil {
		return "", err
	}
	var resp response
	if err = s.readJSON(&resp); err != nil {
		return "", err
	}
	if err = responseError(resp); err != nil {
		return "", err
	}
	return resp.EnvPath, nil
}

func (c *client) ConfigureNet(ctx context.Context, spec *netSpec) error {
	s, err := c.call(ctx)
	if err != nil {
		return err
	}
	defer s.conn.Close()

	if err = s.writeJSON(request{
		Op:      opNet,
		Iface:   spec.Iface,
		Addr:    spec.Addr,
		Gateway: spec.Gateway,
		DNS:     spec.DNS,
	}); err != nil {
		return err
	}
	var resp response
	if err = s.readJSON(&resp); err != nil {
		return err
	}
	return responseError(resp)
}

func (c *client) Stat(ctx context.Context, path string) (fs.FileInfo, error) {
	s, err := c.call(ctx)
	if err != nil {
		return nil, err
	}
	defer s.conn.Close()

	if err = s.writeJSON(request{Op: opStat, Path: path}); err != nil {
		return nil, err
	}
	var resp response
	if err = s.readJSON(&resp); err != nil {
		return nil, err
	}
	if err = responseError(resp); err != nil {
		return nil, err
	}
	return fileInfoFrom(resp), nil
}

func (c *client) CopyIn(ctx context.Context, reader io.Reader, dst string) error {
	s, err := c.call(ctx)
	if err != nil {
		return err
	}
	defer s.conn.Close()

	if err = s.writeJSON(request{Op: opCopyIn, Path: dst}); err != nil {
		return err
	}
	if _, err = io.Copy(s.conn, reader); err != nil {
		return err
	}
	var resp response
	if err = s.readJSON(&resp); err != nil {
		return err
	}
	return responseError(resp)
}

func (c *client) CopyOut(ctx context.Context, src string) (io.ReadCloser, error) {
	s, err := c.call(ctx)
	if err != nil {
		return nil, err
	}

	if err = s.writeJSON(request{Op: opCopyOut, Path: src}); err != nil {
		s.conn.Close()
		return nil, err
	}
	var resp response
	if err = s.readJSON(&resp); err != nil {
		s.conn.Close()
		return nil, err
	}
	if err = responseError(resp); err != nil {
		s.conn.Close()
		return nil, err
	}
	return &tarReadCloser{Reader: s.r, closer: s.conn}, nil
}

func (c *client) Execute(ctx context.Context, cmd, path []string, env map[string]string, workdir string, streams *stream.Streams) error {
	s, err := c.call(ctx)
	if err != nil {
		return err
	}
	defer s.conn.Close()

	if err = s.writeJSON(request{
		Op:      opExec,
		Cmd:     cmd,
		Env:     env,
		PathEnv: path,
		Workdir: workdir,
	}); err != nil {
		return err
	}
	var resp response
	if err = s.readJSON(&resp); err != nil {
		return err
	}
	if err = responseError(resp); err != nil {
		return err
	}

	go func() {
		if streams != nil && streams.In != nil {
			buf := make([]byte, 32*1024)
			for {
				n, err := streams.In.Read(buf)
				if n > 0 {
					if werr := s.writeFrame(streamStdin, buf[:n]); werr != nil {
						return
					}
				}
				if err != nil {
					_ = s.writeFrame(streamStdin, nil)
					return
				}
			}
		} else {
			_ = s.writeFrame(streamStdin, nil)
		}
	}()

	stdout, stderr := io.Discard, io.Discard
	if streams != nil {
		if streams.Out != nil {
			stdout = streams.Out
		}
		if streams.Err != nil {
			stderr = streams.Err
		}
	}

	for {
		streamID, payload, err := s.readFrame()
		if err != nil {
			return err
		}
		switch streamID {
		case streamStdout:
			if _, err = stdout.Write(payload); err != nil {
				return err
			}
		case streamStderr:
			if _, err = stderr.Write(payload); err != nil {
				return err
			}
		case streamExit:
			if len(payload) < 4 {
				return fmt.Errorf("invalid exit frame")
			}
			code := int(binary.BigEndian.Uint32(payload[:4]))
			if code != 0 {
				return fmt.Errorf("exitcode '%v': failure", code)
			}
			return nil
		}
	}
}

type tarReadCloser struct {
	io.Reader
	closer io.Closer
}

func (t *tarReadCloser) Close() error {
	return t.closer.Close()
}
