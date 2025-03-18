/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package traits

import (
	"iter"

	"drassi.run/core/pkg/expression/types/ref"
)

// Iterable aggregate types permit traversal over their elements.
type Iterable interface {
	// Items return a new iterator view of the struct.
	Items() Iterator
}

// Iterator permits safe traversal over the contents of an aggregate type.
type Iterator iter.Seq2[ref.Val, ref.Val]
