/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package libraries

import (
	"drassi.run/core/pkg/expression/types"
	"testing"
)

func TestStartsWith(t *testing.T) {
	type testCase struct {
		str    any
		prefix any
		result bool
	}
	tests := []testCase{
		{"search", "se", true},
		{"search", "SE", true},
		{"search", "sa", false},
		{"123search", "123s", true},
		{12, "s", false},
		{12, "12", true},
		{nil, "42", false},
		{"null", nil, true},
		{"null", "", true},
		{listInt, "", false},
		{mapSS, "", false},
		{objectX, "", false},
	}
	for _, test := range tests {
		str := types.NativeToVal(test.str)
		prefix := types.NativeToVal(test.prefix)
		actual := StartsWith(str, prefix)

		verify(t, test.result, actual, "startsWith(%v, %v)", test.str, test.prefix)
	}
}

func TestEndsWith(t *testing.T) {
	type testCase struct {
		str    any
		suffix any
		result bool
	}
	tests := []testCase{
		{"search", "ch", true},
		{"search", "CH", true},
		{"search", "rh", false},
		{"search123s", "123s", true},
		{12, "s", false},
		{12, "12", true},
		{nil, "42", false},
		{"null", nil, true},
		{"null", "", true},
		{listInt, "", false},
		{mapSS, "", false},
		{objectX, "", false},
	}
	for _, test := range tests {
		str := types.NativeToVal(test.str)
		suffix := types.NativeToVal(test.suffix)
		actual := EndsWith(str, suffix)

		verify(t, test.result, actual, "endsWith(%v, %v)", test.str, test.suffix)
	}
}
