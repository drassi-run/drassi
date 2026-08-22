/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package firecracker

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"sync"
	"time"
)

// Protocol is a JSON line followed by an optional payload on a vsock stream.
//
// Kata talks ttrpc CopyFile over Firecracker hybrid vsock (host Unix socket +
// "CONNECT <port>\n"). That RPC is host→guest only and chunks each file.
// Drassi CopyIn/CopyOut already speak tar, so we stream tar on the same
// hybrid-vsock channel instead. The guest stops at the tar end marker;
// vsock stays open for the JSON ack (do not CloseWrite — Firecracker
// hybrid vsock treats a shutdown as a full close).
const (
	opCopyIn  = "copyin"
	opCopyOut = "copyout"
	opStat    = "stat"
	opExec    = "exec"
	opInfo    = "info"
	opNet     = "net"

	codeNotExist = "not_exist"

	streamStdin  byte = 0
	streamStdout byte = 1
	streamStderr byte = 2
	streamExit   byte = 3

	defaultGuestCID  = 3
	defaultAgentPort = 1024
)

type request struct {
	Op      string            `json:"op"`
	Path    string            `json:"path,omitempty"`
	Cmd     []string          `json:"cmd,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	PathEnv []string          `json:"path_env,omitempty"`
	Workdir string            `json:"workdir,omitempty"`
	Iface   string            `json:"iface,omitempty"`
	Addr    string            `json:"addr,omitempty"`
	Gateway string            `json:"gateway,omitempty"`
	DNS     []string          `json:"dns,omitempty"`
}

type response struct {
	OK      bool   `json:"ok"`
	Error   string `json:"error,omitempty"`
	Code    string `json:"code,omitempty"`
	Name    string `json:"name,omitempty"`
	Size    int64  `json:"size,omitempty"`
	Mode    uint32 `json:"mode,omitempty"`
	ModTime int64  `json:"mod_time,omitempty"`
	EnvPath string `json:"env_path,omitempty"`
}

func okResponse() response {
	return response{OK: true}
}

func errResponse(err error) response {
	resp := response{Error: err.Error()}
	if errors.Is(err, fs.ErrNotExist) {
		resp.Code = codeNotExist
	}
	return resp
}

func responseError(resp response) error {
	if resp.OK {
		return nil
	}
	if resp.Code == codeNotExist {
		return fmt.Errorf("%s: %w", resp.Error, fs.ErrNotExist)
	}
	if resp.Error == "" {
		return fmt.Errorf("agent request failed")
	}
	return fmt.Errorf("%s", resp.Error)
}

type session struct {
	conn net.Conn
	r    *bufio.Reader
	wmu  sync.Mutex
}

func newSession(conn net.Conn) *session {
	return &session{
		conn: conn,
		r:    bufio.NewReader(conn),
	}
}

func (s *session) readJSON(v any) error {
	line, err := s.r.ReadBytes('\n')
	if err != nil {
		return err
	}
	return json.Unmarshal(line, v)
}

func (s *session) writeJSON(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	s.wmu.Lock()
	defer s.wmu.Unlock()
	_, err = s.conn.Write(append(b, '\n'))
	return err
}

func (s *session) writeFrame(stream byte, p []byte) error {
	var hdr [5]byte
	hdr[0] = stream
	binary.BigEndian.PutUint32(hdr[1:], uint32(len(p)))

	s.wmu.Lock()
	defer s.wmu.Unlock()
	if _, err := s.conn.Write(hdr[:]); err != nil {
		return err
	}
	if len(p) == 0 {
		return nil
	}
	_, err := s.conn.Write(p)
	return err
}

func (s *session) readFrame() (byte, []byte, error) {
	var hdr [5]byte
	if _, err := io.ReadFull(s.r, hdr[:]); err != nil {
		return 0, nil, err
	}
	n := binary.BigEndian.Uint32(hdr[1:])
	if n == 0 {
		return hdr[0], nil, nil
	}
	p := make([]byte, n)
	if _, err := io.ReadFull(s.r, p); err != nil {
		return 0, nil, err
	}
	return hdr[0], p, nil
}

type fileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
}

func fileInfoFrom(resp response) *fileInfo {
	return &fileInfo{
		name:    resp.Name,
		size:    resp.Size,
		mode:    fs.FileMode(resp.Mode),
		modTime: time.Unix(0, resp.ModTime),
	}
}

func (fi *fileInfo) Name() string       { return fi.name }
func (fi *fileInfo) Size() int64        { return fi.size }
func (fi *fileInfo) Mode() fs.FileMode  { return fi.mode }
func (fi *fileInfo) ModTime() time.Time { return fi.modTime }
func (fi *fileInfo) IsDir() bool        { return fi.mode.IsDir() }
func (fi *fileInfo) Sys() any           { return nil }
