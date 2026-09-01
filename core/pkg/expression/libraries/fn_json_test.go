/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package libraries

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"testing"

	"drassi.run/core/pkg/expression/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var jsonTest = [][2]any{
	{nil, "null"},
	{true, "true"},
	{false, "false"},
	{0, "0"},
	{-1, "-1"},
	{1, "1"},
	{0.0, "0"},
	{3.14, "3.14"},
	{-3.1, "-3.1"},
	{"", `""`},
	{"0", `"0"`},
	{"-1", `"-1"`},
	{"1", `"1"`},
	{"-Infinity", `"-Infinity"`},
	{"Infinity", `"Infinity"`},
	{"NaN", `"NaN"`},
	{"foobar", `"foobar"`},

	//{math.Inf(-1), "-Infinity"},
	//{math.Inf(1), "Infinity"},
	//{math.NaN(), "NaN"},
	//{[]any{"first", math.Inf(-1), math.Inf(1), math.NaN()}, nil},
	//{map[string]any{"first": "one", "second": math.Inf(-1), "third": math.Inf(1), "fourth": math.NaN()}, nil},
}

func TestToJSON(t *testing.T) {
	t.Run("primitive", testToJSONPrimitive)
	t.Run("non-primitive", testToJSONNonPrimitive)
}

func testToJSONPrimitive(t *testing.T) {
	for _, tc := range jsonTest {
		actual := ToJSON(types.NativeToVal(tc[0]))
		verify(t, tc[1], actual, "toJSON(%v)", tc[0])
	}
}

func testToJSONNonPrimitive(t *testing.T) {
	tests := []any{listInt, listString, listFloat, mapSS, mapIS, objectX}
	for _, tc := range tests {
		expected, _ := json.Marshal(tc, jsontext.WithIndent("  "))
		actual := ToJSON(types.NativeToVal(tc))

		err, _ := actual.(error)
		require.NoError(t, err, "toJSON(%v)", tc)

		assert.JSONEqf(t, string(expected), actual.Value().(string), "toJSON(%v)", tc)
	}
}

func TestFromJSON(t *testing.T) {
	t.Run("primitive", testFromJSONPrimitive)
	t.Run("non-primitive", testFromJSONNonPrimitive)
}

func testFromJSONPrimitive(t *testing.T) {
	for _, tc := range jsonTest {
		actual := FromJson(types.NativeToVal(tc[1]))
		verify(t, tc[0], actual, "fromJSON(%v)", tc[1])
	}
}

func testFromJSONNonPrimitive(t *testing.T) {
	tests := []any{listInt, listString, listFloat, mapSS, mapIS, objectX}

	for _, tc := range tests {
		str, _ := json.Marshal(tc)
		actual := FromJson(types.NativeToVal(str))

		err, _ := actual.(error)
		require.NoError(t, err, "fromJSON(%v)", tc)

		rActual, err := json.Marshal(actual.Value())
		require.NoError(t, err)
		assert.JSONEq(t, string(str), string(rActual), "fromJSON(%v)", tc)
	}
}
