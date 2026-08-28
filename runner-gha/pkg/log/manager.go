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

func (m *Manager) Start(uid string, attrs ...attribute.KeyValue) error {
	m.mu.Lock()
	if _, exists := m.sessions[uid]; exists {
		m.mu.Unlock()
		return fmt.Errorf("session %s already started", uid)
	}

	s := newSession(m.dir, uid, attrs...)
	m.sessions[uid] = s
	m.mu.Unlock()

	m.notify(&Event{
		Uid:   uid,
		Kind:  OnRecordStart,
		Attrs: attrs,
	})
	return nil
}

func (m *Manager) Handle(uid string, line string) error {
	m.mu.RLock()
	s, ok := m.sessions[uid]
	m.mu.RUnlock()

	if !ok {
		return nil
	}

	update, err := s.Write(line, m.maxSize)
	if err != nil {
		return err
	}

	m.notify(&Event{
		Uid:    s.uid,
		Kind:   OnRecordLog,
		Attrs:  s.attrs,
		Update: update,
	})
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

	update, err := s.Stop()
	m.notify(&Event{
		Uid:    s.uid,
		Kind:   OnRecordStop,
		Attrs:  s.attrs,
		Update: update,
	})
	return err
}

func (m *Manager) Close() error {
	m.mu.Lock()
	sessions, subs := m.sessions, m.subs
	m.sessions, m.subs = make(map[string]*session), nil
	m.mu.Unlock()

	for _, s := range sessions {
		_ = s.Close()
	}

	for _, sub := range subs {
		close(sub)
	}
	return nil
}

func (m *Manager) Dispose() error {
	return os.RemoveAll(m.dir)
}

func (m *Manager) notify(e *Event) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, sub := range m.subs {
		sub <- e
	}
}
