/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package messages

import (
	"github.com/google/go-cmp/cmp"
	"gotest.tools/v3/assert"
	"maps"
	"reflect"
	"testing"
)

var opt cmp.Option

func init() {
	float64Type := reflect.TypeFor[float64]()
	isNumeric := func(v any) bool {
		return v != nil && reflect.TypeOf(v).ConvertibleTo(float64Type)
	}

	trans := cmp.Transformer("float64", func(v any) float64 {
		return reflect.ValueOf(v).
			Convert(float64Type).
			Float()
	})
	opt = cmp.FilterValues(func(x, y any) bool {
		return isNumeric(x) && isNumeric(y)
	}, trans)
}

func TestDecodeContextData(t *testing.T) {
	t.Run("nil", func(tt *testing.T) {
		testDecodeContextDataSimpleInput(tt, nil)
	})

	t.Run("bool", func(tt *testing.T) {
		testDecodeContextDataSimpleInput(tt, true)
	})

	t.Run("string", func(tt *testing.T) {
		testDecodeContextDataSimpleInput(tt, "hello world")
	})

	t.Run("number", func(tt *testing.T) {
		testDecodeContextDataSimpleInput(tt, 1234)
		testDecodeContextDataSimpleInput(tt, 1234.0)
	})

	array := []any{
		true, false,
		1234, 43.21,
		"some string",
	}
	dict := map[string]any{
		"t": float64(2),
		"d": []map[string]any{
			{"k": "bool1", "v": true},
			{"k": "bool2", "v": false},
			{"k": "int", "v": 1234},
			{"k": "float", "v": 43.21},
			{"k": "string", "v": "some string"},
		},
	}
	dictexpected := map[string]any{
		"bool1":  true,
		"bool2":  false,
		"int":    1234,
		"float":  43.21,
		"string": "some string",
	}

	t.Run("array/simple", func(tt *testing.T) {
		testDecodeContextDataSimpleInput(tt, array)
	})
	t.Run("array/object", func(tt *testing.T) {
		input := map[string]any{
			"t": float64(1),
			"a": array,
		}
		testDecodeContextDataObject(tt, input, array)
	})
	t.Run("array/simple_with_dict", func(tt *testing.T) {
		input := append(array, dict)
		expected := append(array, dictexpected)
		testDecodeContextDataObject(tt, input, expected)
	})
	t.Run("array/object_with_dict", func(tt *testing.T) {
		input := map[string]any{
			"t": float64(1),
			"a": append(array, dict),
		}
		expected := append(array, dictexpected)
		testDecodeContextDataObject(tt, input, expected)
	})
	t.Run("dict/simple", func(tt *testing.T) {
		testDecodeContextDataObject(tt, dict, dictexpected)
	})
	t.Run("dict/complex", func(tt *testing.T) {
		input := map[string]any{
			"t": float64(2),
			"d": append(
				dict["d"].([]map[string]any),
				map[string]any{
					"k": "map",
					"v": array,
				},
				map[string]any{
					"k": "dict",
					"v": dict,
				},
			),
		}
		expected := maps.Clone(dictexpected)
		expected["map"] = array
		expected["dict"] = dictexpected
		testDecodeContextDataObject(tt, input, expected)
	})
}

func testDecodeContextDataSimpleInput(tt *testing.T, input any) {
	output, err := DecodeContextData("", input)

	assert.NilError(tt, err)
	assert.DeepEqual(tt, input, output, opt)
}

func testDecodeContextDataObject(tt *testing.T, input any, expected any) {
	output, err := DecodeContextData("", input)

	assert.NilError(tt, err)
	assert.DeepEqual(tt, expected, output, opt)
}
