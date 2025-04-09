/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package reactive

import (
	"sync"
)

type WaitState[T comparable] struct {
	mu    sync.Mutex
	cond  *sync.Cond
	state T
}

func NewWaitState[T comparable](initState T) *WaitState[T] {
	ws := &WaitState[T]{
		state: initState,
	}
	ws.cond = sync.NewCond(&ws.mu)
	return ws
}

func (ws *WaitState[T]) Set(state T) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.state = state
	ws.cond.Broadcast()
}

func (ws *WaitState[T]) Get() T {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	return ws.state
}

func (ws *WaitState[T]) Wait(target T) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	for ws.state != target {
		ws.cond.Wait()
	}
}
