/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// RFC3339Tick is C# DateTime format with precision of 100 nanoseconds (7 digits)
const RFC3339Tick = "2006-01-02T15:04:05.0000000Z07:00"

type session struct {
	dir   string
	uid   string
	attrs []attribute.KeyValue

	mu    sync.Mutex
	f     *os.File
	idx   int
	size  int64
	lines int
}

func newSession(dir string, uid string, attrs ...attribute.KeyValue) *session {
	return &session{
		dir:   dir,
		uid:   uid,
		attrs: attrs,
	}
}

func (s *session) filePath() string {
	return filepath.Join(s.dir, fmt.Sprintf("%s.%d.log", s.uid, s.idx))
}

func (s *session) ensureFile() error {
	if s.f != nil {
		return nil
	}
	f, err := os.Create(s.filePath())
	if err != nil {
		return err
	}
	s.f = f
	s.size = 0
	s.lines = 0
	return nil
}

func (s *session) writeLine(line string) error {
	now := time.Now().UTC()
	b := now.AppendFormat(nil, RFC3339Tick)
	b = append(b, ' ')
	b = append(b, line...)
	if l := len(line); l == 0 || line[l-1] != '\n' {
		b = append(b, '\n')
	}

	n, err := s.f.Write(b)
	if err != nil {
		return err
	}

	s.size += int64(n)
	s.lines++
	return nil
}

func (s *session) rotate() error {
	if s.f == nil {
		return nil
	}
	_ = s.f.Chmod(0400)
	if err := s.f.Close(); err != nil {
		return err
	}

	s.idx++
	s.f = nil
	s.size = 0
	s.lines = 0
	return nil
}

func (s *session) Write(line string, maxSize int64) (*Update, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureFile(); err != nil {
		return nil, err
	}

	if err := s.writeLine(line); err != nil {
		return nil, err
	}

	u := &Update{
		File:     s.filePath(),
		Complete: false,
		Offset:   s.size,
		Line:     s.lines,
	}

	if s.size >= maxSize {
		u.Complete = true
		if err := s.rotate(); err != nil {
			return nil, err
		}
	}

	return u, nil
}

func (s *session) Stop() (*Update, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.f == nil {
		return nil, nil
	}

	u := &Update{
		File:     s.filePath(),
		Complete: true,
		Offset:   s.size,
		Line:     s.lines,
	}

	return u, s.rotate()
}

func (s *session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.rotate()
}
