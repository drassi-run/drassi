/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package log

import (
	"context"
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
		dir:     dir,
		maxSize: maxSize,
	}
	return m, nil
}

// Manager writes log entries to the local filesystem then
// dispatches Update Event to active subscribers.
// Rotation is performed automatically when file size thresholds are met.
type Manager struct {
	dir     string
	maxSize int64

	currUid   string
	currAttrs []attribute.KeyValue
	currSize  int64
	currLines int
	idx       int
	f         *os.File

	mu   sync.Mutex
	subs []chan<- *Event
}

func (m *Manager) Subscribe() <-chan *Event {
	ch := make(chan *Event, 5)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.subs = append(m.subs, ch)
	return ch
}

// ContextHandle is used for [scribe.Handler]
func (m *Manager) ContextHandle(_ context.Context, msg string) error {
	return m.Handle(msg)
}

func (m *Manager) Handle(line string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.f == nil {
		path := m.currFile()
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		m.f, m.currSize, m.currLines = f, 0, 0
	}

	// write data
	if err := m.write(line); err != nil {
		return err
	}

	e := &Event{
		Uid:   m.currUid,
		Kind:  OnRecordLog,
		Attrs: m.currAttrs,
		Update: &Update{
			File:     m.currFile(),
			Complete: false,
			Offset:   m.currSize,
			Line:     m.currLines,
		},
	}

	if m.currSize >= m.maxSize {
		e.Complete = true
		if err := m.rotate(); err != nil {
			return err
		}
	}

	m.notify(e)
	return nil
}

func (m *Manager) currFile() string {
	return fmt.Sprintf("%s/%s.%d.log", m.dir, m.currUid, m.idx)
}

// RFC3339Tick is C# DateTime format with precision of 100 nanoseconds (7 digits)
const RFC3339Tick = "2006-01-02T15:04:05.0000000Z07:00"

func (m *Manager) write(line string) error {
	now := time.Now().UTC()
	b := now.AppendFormat(nil, RFC3339Tick)
	b = append(b, ' ')
	b = append(b, line...)
	if l := len(line); l == 0 || line[l-1] != '\n' {
		b = append(b, '\n')
	}

	n, err := m.f.Write(b)
	if err != nil {
		return err
	}

	m.currSize += int64(n)
	m.currLines++
	return nil
}

func (m *Manager) rotate() error {
	// Chmod to RO
	_ = m.f.Chmod(0400)
	if err := m.f.Close(); err != nil {
		return err
	}

	m.f, m.currLines, m.currSize = nil, 0, 0
	m.idx++
	return nil
}

func (m *Manager) notify(e *Event) {
	for _, sub := range m.subs {
		sub <- e
	}
}

func (m *Manager) Start(uid string, attrs ...attribute.KeyValue) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currUid != "" {
		return fmt.Errorf("session already started")
	}

	m.currUid, m.currAttrs, m.idx = uid, attrs, 0
	m.f, m.currLines, m.currSize = nil, 0, 0

	e := &Event{
		Uid:   m.currUid,
		Kind:  OnRecordStart,
		Attrs: m.currAttrs,
	}
	m.notify(e)
	return nil
}

func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currUid == "" {
		return nil
	}

	e := &Event{
		Uid:   m.currUid,
		Kind:  OnRecordStop,
		Attrs: m.currAttrs,
	}

	if m.f != nil {
		e.Update = &Update{
			File:     m.currFile(),
			Complete: true,
			Offset:   m.currSize,
			Line:     m.currLines,
		}

		if err := m.rotate(); err != nil {
			return err
		}
	}

	m.currUid, m.currAttrs, m.idx = "", nil, 0
	m.notify(e)
	return nil
}

func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.currUid, m.currAttrs, m.idx = "", nil, 0

	for _, sub := range m.subs {
		close(sub)
	}
	m.subs = nil // avoid panic when close twice
	return nil
}

func (m *Manager) Dispose() error {
	return os.RemoveAll(m.dir)
}
