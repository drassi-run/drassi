/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package evaluator

import (
	"iter"
	"testing"

	"drassi.run/core/pkg/model"
	. "drassi.run/core/pkg/model/workflows"
	"github.com/stretchr/testify/assert"
)

type Tuple2[T1, T2 any] struct {
	V1 T1
	V2 T2
}

func iterOfTuple2[T1, T2 any](t []Tuple2[T1, T2]) iter.Seq2[T1, T2] {
	return func(yield func(T1, T2) bool) {
		for _, v := range t {
			if !yield(v.V1, v.V2) {
				return
			}
		}
	}
}

func iterOfMap[K comparable, V any](m map[K]V) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range m {
			if !yield(k, v) {
				return
			}
		}
	}
}

func evaluateSuccess[R any](evals iter.Seq2[Evaluable[R], R]) func(t *testing.T) {
	return func(t *testing.T) {
		for e, expected := range evals {
			actual := new(R)
			err := Evaluate(env, e, actual)
			assert.NoError(t, err)
			assert.Equal(t, expected, *actual)
		}
	}
}

func evaluateFailed[R any](evals []Evaluable[R], contains string) func(t *testing.T) {
	return func(t *testing.T) {
		for _, e := range evals {
			o := new(R)
			err := Evaluate(env, e, o)
			assert.ErrorContains(t, err, contains, "%#v", e)
		}
	}
}

func tokenFrom(input any) Token {
	var t Token
	if err := model.Decode(input, &t); err != nil {
		panic(err)
	}
	return t
}

func TestEvaluate(t *testing.T) {
	t.Run("string", testEvaluateString)
	t.Run("bool", testEvaluateBool)
	t.Run("int64", testEvaluateInt64)
	t.Run("float64", testEvaluateFloat64)
	t.Run("list", testEvaluateList)
	t.Run("map", testEvaluateMap)
	t.Run("container", testEvaluateContainer)
	t.Run("runs-on", testEvaluateRunsOn)
}

func testEvaluateString(t *testing.T) {
	tcSuccess := map[Evaluable[string]]string{
		NewLiteralToken("foobar"):             "foobar",
		NewLiteralToken(123):                  "123",
		NewLiteralToken(3.14):                 "3.14",
		NewLiteralToken(true):                 "true",
		NewExpressionToken("${{ 'foobar' }}"): "foobar",
	}

	t.Run("success", evaluateSuccess(iterOfMap(tcSuccess)))

	tcFailed := []Evaluable[string]{
		NewExpressionToken("${{ la }}"),
		NewExpressionToken("${{ ms }}"),
	}
	t.Run("failed", evaluateFailed(tcFailed, "got unconvertible type"))
}

func testEvaluateBool(t *testing.T) {
	tcSuccess := map[Evaluable[bool]]bool{
		NewLiteralToken(true):              true,
		NewExpressionToken("${{ false }}"): false,
	}

	t.Run("success", evaluateSuccess(iterOfMap(tcSuccess)))

	tcFailed := []Evaluable[bool]{
		NewLiteralToken("foobar"),
		NewLiteralToken(123),
		NewLiteralToken(3.14),
		NewExpressionToken("${{ la }}"),
		NewExpressionToken("${{ ms }}"),
	}
	t.Run("failed", evaluateFailed(tcFailed, "got unconvertible type"))
}

func testEvaluateInt64(t *testing.T) {
	tcSuccess := map[Evaluable[int64]]int64{
		NewLiteralToken(123):              123,
		NewLiteralToken(3.14):             3,
		NewExpressionToken("${{ 123 }}"):  123,
		NewExpressionToken("${{ 3.14 }}"): 3,
	}

	t.Run("success", evaluateSuccess(iterOfMap(tcSuccess)))

	tcFailed := []Evaluable[int64]{
		NewLiteralToken("foobar"),
		NewLiteralToken(true),
		NewExpressionToken("${{ la }}"),
		NewExpressionToken("${{ ms }}"),
	}
	t.Run("failed", evaluateFailed(tcFailed, "got unconvertible type"))
}

func testEvaluateFloat64(t *testing.T) {
	tcSuccess := map[Evaluable[float64]]float64{
		NewLiteralToken(123):              123,
		NewLiteralToken(3.14):             3.14,
		NewExpressionToken("${{ 123 }}"):  123,
		NewExpressionToken("${{ 3.14 }}"): 3.14,
	}

	t.Run("success", evaluateSuccess(iterOfMap(tcSuccess)))

	tcFailed := []Evaluable[float64]{
		NewLiteralToken("foobar"),
		NewLiteralToken(true),
		NewExpressionToken("${{ la }}"),
		NewExpressionToken("${{ ms }}"),
	}
	t.Run("failed", evaluateFailed(tcFailed, "got unconvertible type"))
}

var listResultString = []string{
	"string", "123",
	"abc", "true", "3.14", // la
	"Infinity", "true",
	"one", "two", "three", // ls
}

func testEvaluateList(t *testing.T) {
	tcString := []Tuple2[Evaluable[[]string], []string]{
		{listToken, listResultString},
		{NewExpressionToken(`${{ fromJson('["string",123,true]') }}`), []string{"string", "123", "true"}},
	}

	t.Run("string", evaluateSuccess(iterOfTuple2(tcString)))

	tcInt := []Tuple2[Evaluable[[]int], []int]{
		{
			NewSequenceToken([]Token{
				NewLiteralToken(3.14),
				NewLiteralToken(123),
				NewExpressionToken(`${{ 0 }}`),
			}),
			[]int{3, 123, 0},
		},
		{NewExpressionToken(`${{ fromJson('[3.14,123,0]') }}`), []int{3, 123, 0}},
	}
	t.Run("int", evaluateSuccess(iterOfTuple2(tcInt)))

	tcFloat := []Tuple2[Evaluable[[]float64], []float64]{
		{
			NewSequenceToken([]Token{
				NewLiteralToken(3.14),
				NewLiteralToken(123),
				NewExpressionToken(`${{ 0 }}`),
			}),
			[]float64{3.14, 123, 0},
		},
		{NewExpressionToken(`${{ fromJson('[3.14,123,0]') }}`), []float64{3.14, 123, 0}},
	}
	t.Run("float", evaluateSuccess(iterOfTuple2(tcFloat)))

	tcAny := []Tuple2[Evaluable[[]any], []any]{
		{listToken, listResult},
		{NewExpressionToken(`${{ fromJson('["string",123,true]') }}`), []any{"string", float64(123), true}},
	}
	t.Run("any", evaluateSuccess(iterOfTuple2(tcAny)))

	tcFailed := []Evaluable[[]string]{
		NewLiteralToken("foobar"),
		NewLiteralToken(123),
		NewLiteralToken(3.14),
		NewLiteralToken(true),
		NewExpressionToken("${{ 'foobar' }}"),
		NewExpressionToken("${{ ms }}"),
	}
	t.Run("failed", evaluateFailed(tcFailed, "source data must be an array or slice"))
}

var mapResultString = map[string]string{
	"string":   "foobar",
	"123":      "123",
	"3.14":     "3.14",
	"false":    "true",
	"expr-key": "expr-value",
	"1":        "value", "2": "3.14", "3": "false", // mi
}

func testEvaluateMap(t *testing.T) {
	tcAny := []Tuple2[Evaluable[map[string]any], map[string]any]{
		{mapToken, mapResult},
		{
			NewExpressionToken(`${{ fromJson('{"first":"one","second":2,"third":true}') }}`),
			map[string]any{"first": "one", "second": float64(2), "third": true},
		},
	}
	t.Run("any", evaluateSuccess(iterOfTuple2(tcAny)))

	tcString := []Tuple2[Evaluable[map[string]string], map[string]string]{
		{mapToken, mapResultString},
		{
			NewExpressionToken(`${{ fromJson('{"first":"one","second":2,"third":true}') }}`),
			map[string]string{"first": "one", "second": "2", "third": "true"},
		},
	}
	t.Run("string", evaluateSuccess(iterOfTuple2(tcString)))

	tcFailed := []Evaluable[map[string]any]{
		NewLiteralToken("foobar"),
		NewLiteralToken(123),
		NewLiteralToken(3.14),
		NewLiteralToken(true),
		NewExpressionToken("${{ 'foobar' }}"),
		NewExpressionToken("${{ la }}"),
	}
	t.Run("failed", evaluateFailed(tcFailed, "expected a map"))
}

func testEvaluateContainer(t *testing.T) {
	c1 := map[string]any{
		"image": "foobar",
	}
	c2 := map[string]any{
		"image": "${{ 'foobar' }}",
	}
	c3 := map[string]any{
		"image": "foobar",
		"env": map[string]any{
			"ABC": "XYZ",
		},
	}
	c4 := map[string]any{
		"image": "foobar",
		"env":   `${{ fromJson('{"ABC": "XYZ"}') }}`,
	}
	c5 := map[string]any{
		"image":   "foobar",
		"volumes": []string{"a:a", "b:b"},
	}
	c6 := map[string]any{
		"image":   "foobar",
		"volumes": `${{ fromJson('["a:a", "b:b"]') }}`,
	}
	c7 := `${{ fromJson('{"image":"foobar","env":{"ABC":"XYZ"},"volumes":["a:a","b:b"]}') }}`

	tcSuccess := []Tuple2[Evaluable[*Container], *Container]{
		{NewLiteralToken("foobar"), &Container{Image: "foobar"}},
		{NewLiteralToken(123), &Container{Image: "123"}},
		{NewLiteralToken(3.14), &Container{Image: "3.14"}},
		{NewLiteralToken(true), &Container{Image: "true"}},
		{NewExpressionToken("${{ 'foobar' }}"), &Container{Image: "foobar"}},

		{tokenFrom(c1), &Container{Image: "foobar"}},
		{tokenFrom(c2), &Container{Image: "foobar"}},
		{tokenFrom(c3), &Container{Image: "foobar", Env: map[string]string{"ABC": "XYZ"}}},
		{tokenFrom(c4), &Container{Image: "foobar", Env: map[string]string{"ABC": "XYZ"}}},
		{tokenFrom(c5), &Container{Image: "foobar", Volumes: []string{"a:a", "b:b"}}},
		{tokenFrom(c6), &Container{Image: "foobar", Volumes: []string{"a:a", "b:b"}}},
		{tokenFrom(c7), &Container{Image: "foobar", Env: map[string]string{"ABC": "XYZ"}, Volumes: []string{"a:a", "b:b"}}},
	}

	evaluateSuccess(iterOfTuple2(tcSuccess))(t)
}

func testEvaluateRunsOn(t *testing.T) {
	r1 := []any{
		"foobar",
		"${{ 'abc' }}",
		`${{ fromJson('["first", "second", "third"]') }}`,
	}
	r2 := map[string]any{
		"group":  "ubuntu-runners",
		"labels": "foobar",
	}
	r3 := map[string]any{
		"group": "ubuntu-runners",
		"labels": []any{
			"foobar",
			"${{ 'abc' }}",
			`${{ fromJson('["first", "second", "third"]') }}`,
		},
	}
	tcSuccess := []Tuple2[Evaluable[RunsOn], RunsOn]{
		{NewLiteralToken("foobar"), RunsOn{Labels: []string{"foobar"}}},
		{NewLiteralToken(123), RunsOn{Labels: []string{"123"}}},
		{NewLiteralToken(3.14), RunsOn{Labels: []string{"3.14"}}},
		{NewLiteralToken(true), RunsOn{Labels: []string{"true"}}},
		{NewExpressionToken("${{ 'foobar' }}"), RunsOn{Labels: []string{"foobar"}}},

		{
			tokenFrom(r1),
			RunsOn{Labels: []string{"foobar", "abc", "first", "second", "third"}},
		},
		{
			tokenFrom(r2),
			RunsOn{Group: "ubuntu-runners", Labels: []string{"foobar"}},
		},
		{
			tokenFrom(r3),
			RunsOn{Group: "ubuntu-runners", Labels: []string{"foobar", "abc", "first", "second", "third"}},
		},
	}
	evaluateSuccess(iterOfTuple2(tcSuccess))(t)
}
