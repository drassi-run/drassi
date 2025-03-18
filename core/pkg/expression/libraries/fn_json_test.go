/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package libraries

import (
	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/model"
	"encoding/json"
	"github.com/go-viper/mapstructure/v2"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
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
		actual := ToJSON(types.NativeToVal(tc))
		expected, _ := json.MarshalIndent(tc, "", "  ")
		verify(t, expected, actual, "toJSON(%v)", tc)
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
	tests := []any{objectX} // listInt, listString, listFloat, mapSS, mapIS, objectX}

	for _, tc := range tests {
		str, _ := json.Marshal(tc)
		actual := FromJson(types.NativeToVal(str))

		err, _ := actual.(error)
		assert.NoError(t, err, "fromJSON(%v)", tc)

		rExpected := reflect.ValueOf(tc)
		rActual := reflect.ValueOf(actual.Value())
		assert.True(t, compareUsingReflect(rExpected, rActual), "fromJSON(%v)", tc)
	}
}

var optTagJson = func(config *mapstructure.DecoderConfig) {
	config.TagName = "json"
}

func compareUsingReflect(x, y reflect.Value) bool {
	if !x.IsValid() || !y.IsValid() {
		return x.IsValid() == y.IsValid()
	}

	if kind := x.Kind(); kind == reflect.Struct {
		o := make(map[string]any)
		_ = model.DecodeWithOptions(x.Interface(), &o, optTagJson)
		return compareUsingReflect(reflect.ValueOf(o), y)
	} else if kind == reflect.Pointer || kind == reflect.Interface {
		return compareUsingReflect(x.Elem(), y)
	}

	if kind := y.Kind(); kind == reflect.Struct {
		o := make(map[string]any)
		_ = model.DecodeWithOptions(y.Interface(), &o, optTagJson)
		return compareUsingReflect(x, reflect.ValueOf(o))
	} else if kind == reflect.Pointer || kind == reflect.Interface {
		return compareUsingReflect(x, y.Elem())
	}

	switch x.Kind() {
	case reflect.Array, reflect.Slice:
		if yKind := y.Kind(); yKind != reflect.Array && yKind != reflect.Slice {
			return false
		}
		if x.Len() != y.Len() {
			return false
		}
		for i := 0; i < x.Len(); i++ {
			xItem := x.Index(i)
			yItem := y.Index(i)
			if !compareUsingReflect(xItem, yItem) {
				return false
			}
		}
		return true
	case reflect.Map:
		if y.Kind() != reflect.Map {
			return false
		}
		if x.Len() != y.Len() {
			return false
		}
		for _, k := range y.MapKeys() {
			xItem := y.MapIndex(k)
			yItem := y.MapIndex(k)
			if !compareUsingReflect(xItem, yItem) {
				return false
			}
		}
		return true
	//case reflect.Struct:
	//	if x.Type() != y.Type() {
	//		return false
	//	}
	//	for i := 0; i < x.NumField(); i++ {
	//		xField := x.Field(i)
	//		yField := y.Field(i)
	//		if !compareUsingReflect(xField, yField) {
	//			return false
	//		}
	//	}
	//	return true
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		if x.Kind() != y.Kind() {
			return false
		}
		return x.UnsafePointer() == y.UnsafePointer()
	default:
		xV := x.Interface()
		yV := y.Interface()
		if xV == yV { // same type
			return true
		}

		if !x.CanConvert(y.Type()) {
			return false
		}
		if xC := x.Convert(y.Type()).Interface(); xC != yV {
			return false
		}
		if !y.CanConvert(x.Type()) {
			return false
		}
		if yC := y.Convert(x.Type()).Interface(); yC != xV {
			return false
		}
		return true
	}
}
