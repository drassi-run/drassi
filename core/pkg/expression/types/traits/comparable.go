/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package traits

import "drassi.run/core/pkg/expression/types/ref"

// Comparable interface for ordering comparisons between values in order to
// support '<', '<=', '>=', '>' operators.
type Comparable interface {
	// Compare this value to the input other value, returning an int:
	//
	//    this < other  -> -1
	//    this == other ->  0
	//    this > other  ->  1
	//
	// If the comparison cannot be made or is not supported, an error should
	// be returned.
	Compare(other ref.Val) (int, error)
}
