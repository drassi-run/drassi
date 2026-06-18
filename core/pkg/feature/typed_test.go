/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package feature

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	tests := map[string]struct {
		flags        Flags
		defaultValue string
		want         string
	}{
		"found": {
			flags:        MapFlags{"name": "value"},
			defaultValue: "default",
			want:         "value",
		},
		"missing": {
			flags:        MapFlags{"other": "value"},
			defaultValue: "default",
			want:         "default",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := String(tt.flags, "name", tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBool(t *testing.T) {
	tests := map[string]struct {
		flags        Flags
		defaultValue bool
		want         bool
	}{
		"found true": {
			flags:        MapFlags{"enabled": "true"},
			defaultValue: false,
			want:         true,
		},
		"found false": {
			flags:        MapFlags{"enabled": "false"},
			defaultValue: true,
			want:         false,
		},
		"missing": {
			flags:        MapFlags{},
			defaultValue: true,
			want:         true,
		},
		"invalid": {
			flags:        MapFlags{"enabled": "sometimes"},
			defaultValue: true,
			want:         true,
		},
		"typed flag": {
			flags:        typedFlags{boolValue: false},
			defaultValue: true,
			want:         false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := Bool(tt.flags, "enabled", tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInt64(t *testing.T) {
	tests := map[string]struct {
		flags        Flags
		defaultValue int64
		want         int64
	}{
		"found": {
			flags:        MapFlags{"limit": "42"},
			defaultValue: 7,
			want:         42,
		},
		"missing": {
			flags:        MapFlags{},
			defaultValue: 7,
			want:         7,
		},
		"invalid": {
			flags:        MapFlags{"limit": "4.2"},
			defaultValue: 7,
			want:         7,
		},
		"typed flag": {
			flags:        typedFlags{int64Value: 99},
			defaultValue: 7,
			want:         99,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := Int64(tt.flags, "limit", tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFloat64(t *testing.T) {
	tests := map[string]struct {
		flags        Flags
		defaultValue float64
		want         float64
	}{
		"found": {
			flags:        MapFlags{"ratio": "1.5"},
			defaultValue: 2.5,
			want:         1.5,
		},
		"missing": {
			flags:        MapFlags{},
			defaultValue: 2.5,
			want:         2.5,
		},
		"invalid": {
			flags:        MapFlags{"ratio": "half"},
			defaultValue: 2.5,
			want:         2.5,
		},
		"typed flag": {
			flags:        typedFlags{float64Value: 3.5},
			defaultValue: 2.5,
			want:         3.5,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := Float64(tt.flags, "ratio", tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

type typedFlags struct {
	boolValue    bool
	int64Value   int64
	float64Value float64
}

func (f typedFlags) Flag(string) (string, bool) {
	return "invalid", true
}

func (f typedFlags) Bool(string, bool) bool {
	return f.boolValue
}

func (f typedFlags) Int64(string, int64) int64 {
	return f.int64Value
}

func (f typedFlags) Float64(string, float64) float64 {
	return f.float64Value
}
