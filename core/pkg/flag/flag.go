/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package flag

type Flags interface {
	Flag(key string) (value string, ok bool)
}

type MapFlags map[string]string

func (f MapFlags) Flag(key string) (value string, ok bool) {
	value, ok = f[key]
	return
}
