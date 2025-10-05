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
)

type EventKind uint16

const (
	OnRecordStart EventKind = iota
	OnRecordLog
	OnRecordStop
)

type Event struct {
	Kind     EventKind
	Uid      string // UUID of step/job
	File     string // File location
	Complete bool   // Is File completed or not
	Line     int64  // Line number of log record
	Offset   int64  // Offset at the end of log record
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
	currLines int64
	currSize  int64
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

func (m *Manager) Handle(_ context.Context, line string) error {
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
		Uid:      m.currUid,
		Kind:     OnRecordLog,
		File:     m.currFile(),
		Complete: false,
		Line:     m.currLines,
		Offset:   m.currSize,
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

func (m *Manager) write(line string) error {
	if l := len(line); l == 0 || line[l-1] != '\n' {
		line += "\n"
	}

	n, err := m.f.WriteString(line)
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

func (m *Manager) Start(uid string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currUid != "" {
		return fmt.Errorf("session already started")
	}

	m.currUid, m.idx = uid, 0
	m.f, m.currLines, m.currSize = nil, 0, 0

	e := &Event{
		Uid:  m.currUid,
		Kind: OnRecordStart,
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
		Uid:  m.currUid,
		Kind: OnRecordStop,
	}

	if m.f != nil {
		e.File = m.currFile()
		e.Complete = true
		e.Line = m.currLines
		e.Offset = m.currSize

		if err := m.rotate(); err != nil {
			return err
		}
	}

	m.currUid, m.idx = "", 0
	m.notify(e)
	return nil
}

func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currUid != "" {
		return fmt.Errorf("session %q must stop before close log.Manager", m.currUid)
	}

	for _, sub := range m.subs {
		close(sub)
	}
	m.subs = nil // avoid panic when close twice
	return os.RemoveAll(m.dir)
}
