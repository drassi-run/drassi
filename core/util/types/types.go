/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package xtypes

type Pair[K, V any] struct {
	Key   K
	Value V
}

func NewPair[K, V any](k K, v V) *Pair[K, V] {
	return &Pair[K, V]{k, v}
}

type Unwrapper[T any] interface {
	Unwrap() T
}
