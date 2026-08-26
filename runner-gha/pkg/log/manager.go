/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package log

import (
	"fmt"
	"os"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

type EventKind uint16

const (
	OnRecordStart EventKind = iota
	OnRecordLog
	OnRecordStop
)

type Event struct {
	Uid   string // UUID of step/job
	Kind  EventKind
	Attrs []attribute.KeyValue
	*Update
}

func NewManager(dir string, maxSize int64) (*Manager, error) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}
	m := &Manager{
		dir:      dir,
		maxSize:  maxSize,
		sessions: make(map[string]*session),
	}
	return m, nil
}

type session struct {
	uid   string
	attrs []attribute.KeyValue
	mu    sync.Mutex
	f     *os.File
	idx   int
	size  int64
	lines int
}

// Manager writes log entries to the local filesystem then
// dispatches Update Event to active subscribers.
// Rotation is performed automatically when file size thresholds are met.
type Manager struct {
	dir     string
	maxSize int64

	mu       sync.RWMutex
	subs     []chan<- *Event
	sessions map[string]*session
}

func (m *Manager) Subscribe() <-chan *Event {
	ch := make(chan *Event, 5)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.subs = append(m.subs, ch)
	return ch
}

func (m *Manager) Handle(uid string, line string) error {
	m.mu.RLock()
	s, ok := m.sessions[uid]
	m.mu.RUnlock()

	if !ok {
		return nil
	}

	s.mu.Lock()
	if s.f == nil {
		path := m.sessionFile(s)
		f, err := os.Create(path)
		if err != nil {
			s.mu.Unlock()
			return err
		}
		s.f, s.size, s.lines = f, 0, 0
	}

	// write data
	if err := s.write(line); err != nil {
		s.mu.Unlock()
		return err
	}

	e := &Event{
		Uid:   s.uid,
		Kind:  OnRecordLog,
		Attrs: s.attrs,
		Update: &Update{
			File:     m.sessionFile(s),
			Complete: false,
			Offset:   s.size,
			Line:     s.lines,
		},
	}

	if s.size >= m.maxSize {
		e.Complete = true
		if err := s.rotate(); err != nil {
			s.mu.Unlock()
			return err
		}
	}
	s.mu.Unlock()

	m.notify(e)
	return nil
}

func (m *Manager) sessionFile(s *session) string {
	return fmt.Sprintf("%s/%s.%d.log", m.dir, s.uid, s.idx)
}

// RFC3339Tick is C# DateTime format with precision of 100 nanoseconds (7 digits)
const RFC3339Tick = "2006-01-02T15:04:05.0000000Z07:00"

func (s *session) write(line string) error {
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
	_ = s.f.Chmod(0400)
	if err := s.f.Close(); err != nil {
		return err
	}

	s.f, s.lines, s.size = nil, 0, 0
	s.idx++
	return nil
}

func (m *Manager) notify(e *Event) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, sub := range m.subs {
		sub <- e
	}
}

func (m *Manager) Start(uid string, attrs ...attribute.KeyValue) error {
	m.mu.Lock()
	if _, exists := m.sessions[uid]; exists {
		m.mu.Unlock()
		return fmt.Errorf("session %s already started", uid)
	}

	s := &session{
		uid:   uid,
		attrs: attrs,
	}
	m.sessions[uid] = s
	m.mu.Unlock()

	e := &Event{
		Uid:   uid,
		Kind:  OnRecordStart,
		Attrs: attrs,
	}
	m.notify(e)
	return nil
}

func (m *Manager) Stop(uid string) error {
	m.mu.Lock()
	s, ok := m.sessions[uid]
	if ok {
		delete(m.sessions, uid)
	}
	m.mu.Unlock()

	if !ok {
		return nil
	}

	s.mu.Lock()
	e := &Event{
		Uid:   s.uid,
		Kind:  OnRecordStop,
		Attrs: s.attrs,
	}

	var rotErr error
	if s.f != nil {
		e.Update = &Update{
			File:     m.sessionFile(s),
			Complete: true,
			Offset:   s.size,
			Line:     s.lines,
		}

		rotErr = s.rotate()
	}
	s.mu.Unlock()

	m.notify(e)
	return rotErr
}

func (m *Manager) Close() error {
	m.mu.Lock()
	sessions, subs := m.sessions, m.subs
	m.sessions, m.subs = make(map[string]*session), nil
	m.mu.Unlock()

	for _, s := range sessions {
		s.mu.Lock()
		if s.f != nil {
			_ = s.rotate()
		}
		s.mu.Unlock()
	}

	for _, sub := range subs {
		close(sub)
	}
	return nil
}

func (m *Manager) Dispose() error {
	return os.RemoveAll(m.dir)
}
