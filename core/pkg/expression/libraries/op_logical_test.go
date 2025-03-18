/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package libraries

import (
	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

var logicalRight = []any{true, false, nil, -10, 0, 10, 3.14, 0.0, math.Inf(1), math.Inf(-1), math.NaN(), "", "abc", listInt, mapSS, objectX}
var logicalTruthy = []any{true, -10, 10, math.Inf(1), math.Inf(-1), -3.14, 3.14, "foobar", "-1", "1", "-Infinity", "Infinity", "NaN", listInt, mapSS, objectX}
var logicalFalsy = []any{nil, false, 0, 0.0, math.NaN(), ""}

func verify(t *testing.T, expected any, actual ref.Val, msgAndArgs ...any) {
	err, _ := actual.(error)
	assert.NoError(t, err, msgAndArgs...)

	if en, ok := expected.(float64); ok && math.IsNaN(en) {
		an, ok := actual.Value().(float64)
		assert.True(t, ok && math.IsNaN(an), msgAndArgs...)
		return
	}
	assert.EqualValues(t, expected, actual.Value(), msgAndArgs...)
}

func TestLogicalAnd(t *testing.T) {
	t.Run("simple", testLogicalAndSimple)
}

func testLogicalAndSimple(t *testing.T) {
	for _, l := range logicalTruthy {
		lhs := toLazy(l)
		for _, r := range logicalRight {
			rhs := toLazy(r)
			expected := r
			actual := LogicalAnd(lhs, rhs)
			verify(t, expected, actual, "%v && %v", l, r)
		}
	}

	for _, l := range logicalFalsy {
		lhs := toLazy(l)
		for _, r := range logicalRight {
			rhs := toLazy(r)
			expected := l
			actual := LogicalAnd(lhs, rhs)
			verify(t, expected, actual, "%v && %v", l, r)
		}
	}
}

func TestLogicalOr(t *testing.T) {
	t.Run("simple", testLogicalOrSimple)
}

func testLogicalOrSimple(t *testing.T) {
	for _, l := range logicalTruthy {
		lhs := toLazy(l)
		for _, r := range logicalRight {
			rhs := toLazy(r)
			expected := l
			actual := LogicalOr(lhs, rhs)
			verify(t, expected, actual, "%v || %v", l, r)
		}
	}

	for _, l := range logicalFalsy {
		lhs := toLazy(l)
		for _, r := range logicalRight {
			rhs := toLazy(r)
			expected := r
			actual := LogicalOr(lhs, rhs)
			verify(t, expected, actual, "%v || %v", l, r)
		}
	}
}

func TestLogicalNot(t *testing.T) {
	for _, truthy := range logicalTruthy {
		input := types.NativeToVal(truthy)
		actual := LogicalNot(input)
		verify(t, false, actual, "!%v", truthy)
	}

	for _, truthy := range logicalFalsy {
		input := types.NativeToVal(truthy)
		actual := LogicalNot(input)
		verify(t, true, actual, "!%v", truthy)
	}
}
