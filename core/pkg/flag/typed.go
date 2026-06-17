/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package flag

import "strconv"

func String(flags Flags, key string, defaultValue string) string {
	if v, ok := flags.Flag(key); !ok {
		return defaultValue
	} else {
		return v
	}
}

type boolFlag interface {
	Bool(key string, defaultValue bool) bool
}

func Bool(flags Flags, key string, defaultValue bool) bool {
	if f, ok := flags.(boolFlag); ok {
		return f.Bool(key, defaultValue)
	}

	if v, ok := flags.Flag(key); !ok {
		return defaultValue
	} else if b, err := strconv.ParseBool(v); err != nil {
		return defaultValue
	} else {
		return b
	}
}

type int64Flag interface {
	Int64(key string, defaultValue int64) int64
}

func Int64(flags Flags, key string, defaultValue int64) int64 {
	if f, ok := flags.(int64Flag); ok {
		return f.Int64(key, defaultValue)
	}

	if v, ok := flags.Flag(key); !ok {
		return defaultValue
	} else if i, err := strconv.ParseInt(v, 10, 64); err != nil {
		return defaultValue
	} else {
		return i
	}
}

type float64Flag interface {
	Float64(key string, defaultValue float64) float64
}

func Float64(flags Flags, key string, defaultValue float64) float64 {
	if f, ok := flags.(float64Flag); ok {
		return f.Float64(key, defaultValue)
	}

	if v, ok := flags.Flag(key); !ok {
		return defaultValue
	} else if f, err := strconv.ParseFloat(v, 64); err != nil {
		return defaultValue
	} else {
		return f
	}
}
