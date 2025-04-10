/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package xerror

import "fmt"

func Recover(err *error) {
	switch r := recover().(type) {
	case nil:
		return
	case error:
		*err = r
	default:
		*err = fmt.Errorf("panic: %v", r)
	}
}

// Refine wraps err with fmt.Errorf if err is non nil.
// Intended for use with defer and a named error return.
// Inspired by https://github.com/golang/go/issues/32676.
func Refine(err *error, f string, v ...any) {
	if *err != nil {
		*err = fmt.Errorf(f+": %w", append(v, *err)...)
	}
}
