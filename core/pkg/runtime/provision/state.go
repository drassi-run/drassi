/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package provision

import (
	"fmt"
	"sync"

	ocistore "drassi.run/core/pkg/store/oci"
)

type StateKey[T any] string

const KeyImage = StateKey[ocistore.Image]("image")

type State struct {
	m sync.Map
}

func (s *State) Set[T any](key StateKey[T], val T) {
	s.m.Store(string(key), val)
}

func (s *State) Get[T any](key StateKey[T]) (T, bool) {
	if v, ok := s.m.Load(string(key)); ok {
		if val, ok := v.(T); ok {
			return val, true
		}
	}
	var zero T
	return zero, false
}

func (s *State) MustGet[T any](key StateKey[T]) T {
	if val, ok := s.Get(key); ok {
		return val
	}
	panic(fmt.Sprintf("state key %q not found or type mismatch", string(key)))
}
