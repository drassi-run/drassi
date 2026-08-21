/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package firecracker

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

// DialHybridVsock connects to a Firecracker vsock device the same way Kata does:
// dial the host Unix socket, send "CONNECT <port>\n", wait for "OK ...\n".
//
// See: https://github.com/firecracker-microvm/firecracker/blob/main/docs/vsock.md
func DialHybridVsock(ctx context.Context, udsPath string, port uint32) (net.Conn, error) {
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "unix", udsPath)
	if err != nil {
		return nil, err
	}

	if err = handshake(ctx, conn, port); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return conn, nil
}

func handshake(ctx context.Context, conn net.Conn, port uint32) error {
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
		defer conn.SetDeadline(time.Time{})
	}

	if _, err := fmt.Fprintf(conn, "CONNECT %d\n", port); err != nil {
		return fmt.Errorf("vsock connect: %w", err)
	}

	line, err := readLine(conn)
	if err != nil {
		return fmt.Errorf("vsock handshake: %w", err)
	}
	if !strings.HasPrefix(line, "OK") {
		return fmt.Errorf("vsock handshake: %s", line)
	}
	return nil
}

func readLine(r io.Reader) (string, error) {
	var buf []byte
	b := make([]byte, 1)
	for {
		if _, err := r.Read(b); err != nil {
			return string(buf), err
		}
		if b[0] == '\n' {
			return string(buf), nil
		}
		buf = append(buf, b[0])
		if len(buf) > 256 {
			return string(buf), fmt.Errorf("handshake line too long")
		}
	}
}
