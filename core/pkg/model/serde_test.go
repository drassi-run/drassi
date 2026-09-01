/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package model

import (
	"encoding/json/v2"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type weakTestStruct struct {
	Str   string  `json:"str"`
	Int   int64   `json:"int"`
	Uint  uint32  `json:"uint"`
	Bool  bool    `json:"bool"`
	Float float64 `json:"float"`
}

func TestWeakUnmarshalers(t *testing.T) {
	opts := json.WithUnmarshalers(WeakUnmarshalers())

	t.Run("weaklyString", func(t *testing.T) {
		type stringHolder struct {
			Val string `json:"val"`
		}

		// Null
		var s1 stringHolder
		require.NoError(t, json.Unmarshal([]byte(`{"val": null}`), &s1, opts))
		assert.Equal(t, "", s1.Val)

		// Booleans
		var s2 stringHolder
		require.NoError(t, json.Unmarshal([]byte(`{"val": true}`), &s2, opts))
		assert.Equal(t, "true", s2.Val)

		var s3 stringHolder
		require.NoError(t, json.Unmarshal([]byte(`{"val": false}`), &s3, opts))
		assert.Equal(t, "false", s3.Val)

		// Numbers
		var s4 stringHolder
		require.NoError(t, json.Unmarshal([]byte(`{"val": 123}`), &s4, opts))
		assert.Equal(t, "123", s4.Val)

		var s5 stringHolder
		require.NoError(t, json.Unmarshal([]byte(`{"val": -456}`), &s5, opts))
		assert.Equal(t, "-456", s5.Val)

		var s6 stringHolder
		require.NoError(t, json.Unmarshal([]byte(`{"val": 3.14}`), &s6, opts))
		assert.Equal(t, "3.14", s6.Val)

		// Strings
		var s7 stringHolder
		require.NoError(t, json.Unmarshal([]byte(`{"val": "hello world"}`), &s7, opts))
		assert.Equal(t, "hello world", s7.Val)

		var s8 stringHolder
		require.NoError(t, json.Unmarshal([]byte(`{"val": ""}`), &s8, opts))
		assert.Equal(t, "", s8.Val)

		// Unsupported types (array, object)
		var s9 stringHolder
		assert.Error(t, json.Unmarshal([]byte(`{"val": [1, 2]}`), &s9, opts))

		var s10 stringHolder
		assert.Error(t, json.Unmarshal([]byte(`{"val": {"k": "v"}}`), &s10, opts))
	})

	t.Run("weaklyBool", func(t *testing.T) {
		type boolHolder struct {
			Val bool `json:"val"`
		}

		// Null
		var b1 boolHolder
		require.NoError(t, json.Unmarshal([]byte(`{"val": null}`), &b1, opts))
		assert.False(t, b1.Val)

		// Booleans
		var b2 boolHolder
		require.NoError(t, json.Unmarshal([]byte(`{"val": true}`), &b2, opts))
		assert.True(t, b2.Val)

		var b3 boolHolder
		require.NoError(t, json.Unmarshal([]byte(`{"val": false}`), &b3, opts))
		assert.False(t, b3.Val)

		// Truthy strings
		for _, s := range []string{`"1"`, `"yes"`, `"on"`, `"true"`, `"YES"`, `"True"`, `" ON "`, `" 1 "`} {
			var b boolHolder
			require.NoError(t, json.Unmarshal([]byte(`{"val": `+s+`}`), &b, opts), "failed on %s", s)
			assert.True(t, b.Val, "expected true for %s", s)
		}

		// Falsy strings
		for _, s := range []string{`""`, `"0"`, `"no"`, `"off"`, `"false"`, `"NO"`, `"False"`, `" OFF "`, `" 0 "`, `"   "`} {
			var b boolHolder
			require.NoError(t, json.Unmarshal([]byte(`{"val": `+s+`}`), &b, opts), "failed on %s", s)
			assert.False(t, b.Val, "expected false for %s", s)
		}

		// Invalid strings
		for _, s := range []string{`"invalid"`, `"2"`, `"-1"`, `"maybe"`, `"10"`} {
			var b boolHolder
			assert.Error(t, json.Unmarshal([]byte(`{"val": `+s+`}`), &b, opts), "expected error for %s", s)
		}

		// Unsupported types (number, array, object)
		var b4 boolHolder
		assert.Error(t, json.Unmarshal([]byte(`{"val": 1}`), &b4, opts))

		var b5 boolHolder
		assert.Error(t, json.Unmarshal([]byte(`{"val": [true]}`), &b5, opts))

		var b6 boolHolder
		assert.Error(t, json.Unmarshal([]byte(`{"val": {"k": true}}`), &b6, opts))
	})

	t.Run("weaklyInteger", func(t *testing.T) {
		type allInts struct {
			I8   int8   `json:"i8"`
			I16  int16  `json:"i16"`
			I32  int32  `json:"i32"`
			I64  int64  `json:"i64"`
			Int  int    `json:"int"`
			U8   uint8  `json:"u8"`
			U16  uint16 `json:"u16"`
			U32  uint32 `json:"u32"`
			U64  uint64 `json:"u64"`
			Uint uint   `json:"uint"`
		}

		// Nulls
		var nulls allInts
		input := []byte(`{
			"i8": null, "i16": null, "i32": null, "i64": null, "int": null,
			"u8": null, "u16": null, "u32": null, "u64": null, "uint": null
		}`)
		err := json.Unmarshal(input, &nulls, opts)
		require.NoError(t, err)
		assert.Equal(t, allInts{}, nulls)

		// Empty strings
		var empties allInts
		input = []byte(`{
			"i8": "", "i16": "  ", "i32": "", "i64": "", "int": " ",
			"u8": "", "u16": "", "u32": " ", "u64": "", "uint": ""
		}`)
		err = json.Unmarshal(input, &empties, opts)
		require.NoError(t, err)
		assert.Equal(t, allInts{}, empties)

		// Numbers within range
		var nums allInts
		input = []byte(`{
			"i8": -128, "i16": -32768, "i32": -2147483648, "i64": -9223372036854775808, "int": -100,
			"u8": 255, "u16": 65535, "u32": 4294967295, "u64": 18446744073709551615, "uint": 100
		}`)
		err = json.Unmarshal(input, &nums, opts)
		require.NoError(t, err)
		assert.Equal(t, int8(-128), nums.I8)
		assert.Equal(t, int16(-32768), nums.I16)
		assert.Equal(t, int32(-2147483648), nums.I32)
		assert.Equal(t, int64(-9223372036854775808), nums.I64)
		assert.Equal(t, -100, nums.Int)
		assert.Equal(t, uint8(255), nums.U8)
		assert.Equal(t, uint16(65535), nums.U16)
		assert.Equal(t, uint32(4294967295), nums.U32)
		assert.Equal(t, uint64(18446744073709551615), nums.U64)
		assert.Equal(t, uint(100), nums.Uint)

		// Truncating float numbers and float strings
		var floats allInts
		input = []byte(`{
			"i8": 3.14, "i16": -9.99, "i32": "42.7", "i64": "-100.2", "int": 55.55,
			"u8": 12.99, "u16": "999.9", "u32": 12345.67, "u64": "777.0", "uint": 88.88
		}`)
		err = json.Unmarshal(input, &floats, opts)
		require.NoError(t, err)
		assert.Equal(t, int8(3), floats.I8)
		assert.Equal(t, int16(-9), floats.I16)
		assert.Equal(t, int32(42), floats.I32)
		assert.Equal(t, int64(-100), floats.I64)
		assert.Equal(t, 55, floats.Int)
		assert.Equal(t, uint8(12), floats.U8)
		assert.Equal(t, uint16(999), floats.U16)
		assert.Equal(t, uint32(12345), floats.U32)
		assert.Equal(t, uint64(777), floats.U64)
		assert.Equal(t, uint(88), floats.Uint)

		// Base prefixes (hex/octal strings)
		type hexTest struct {
			Val int `json:"val"`
		}
		var hexVal hexTest
		require.NoError(t, json.Unmarshal([]byte(`{"val": "0x10"}`), &hexVal, opts))
		assert.Equal(t, 16, hexVal.Val)

		// Number overflow checks
		type int8Holder struct {
			Val int8 `json:"val"`
		}
		type uint8Holder struct {
			Val uint8 `json:"val"`
		}
		type int16Holder struct {
			Val int16 `json:"val"`
		}
		type uint16Holder struct {
			Val uint16 `json:"val"`
		}
		type int32Holder struct {
			Val int32 `json:"val"`
		}
		type uint32Holder struct {
			Val uint32 `json:"val"`
		}
		type int64Holder struct {
			Val int64 `json:"val"`
		}
		type uint64Holder struct {
			Val uint64 `json:"val"`
		}

		var dummyI8 int8Holder
		assert.Error(t, json.Unmarshal([]byte(`{"val": 128}`), &dummyI8, opts))
		assert.Error(t, json.Unmarshal([]byte(`{"val": -129}`), &dummyI8, opts))
		assert.Error(t, json.Unmarshal([]byte(`{"val": "128"}`), &dummyI8, opts))
		assert.Error(t, json.Unmarshal([]byte(`{"val": "-129"}`), &dummyI8, opts))

		var dummyU8 uint8Holder
		assert.Error(t, json.Unmarshal([]byte(`{"val": 256}`), &dummyU8, opts))
		assert.Error(t, json.Unmarshal([]byte(`{"val": -1}`), &dummyU8, opts))
		assert.Error(t, json.Unmarshal([]byte(`{"val": "256"}`), &dummyU8, opts))
		assert.Error(t, json.Unmarshal([]byte(`{"val": "-1"}`), &dummyU8, opts))
		assert.Error(t, json.Unmarshal([]byte(`{"val": "-1.5"}`), &dummyU8, opts))

		var dummyI16 int16Holder
		assert.Error(t, json.Unmarshal([]byte(`{"val": 32768}`), &dummyI16, opts))
		assert.Error(t, json.Unmarshal([]byte(`{"val": -32769}`), &dummyI16, opts))

		var dummyU16 uint16Holder
		assert.Error(t, json.Unmarshal([]byte(`{"val": 65536}`), &dummyU16, opts))
		assert.Error(t, json.Unmarshal([]byte(`{"val": -1}`), &dummyU16, opts))

		var dummyI32 int32Holder
		assert.Error(t, json.Unmarshal([]byte(`{"val": 2147483648}`), &dummyI32, opts))
		assert.Error(t, json.Unmarshal([]byte(`{"val": -2147483649}`), &dummyI32, opts))

		var dummyU32 uint32Holder
		assert.Error(t, json.Unmarshal([]byte(`{"val": 4294967296}`), &dummyU32, opts))

		var dummyI64 int64Holder
		assert.Error(t, json.Unmarshal([]byte(`{"val": "1e20"}`), &dummyI64, opts))
		assert.Error(t, json.Unmarshal([]byte(`{"val": "-1e20"}`), &dummyI64, opts))

		var dummyU64 uint64Holder
		assert.Error(t, json.Unmarshal([]byte(`{"val": "1e21"}`), &dummyU64, opts))
		assert.Error(t, json.Unmarshal([]byte(`{"val": "-1"}`), &dummyU64, opts))

		// Parse & Syntax Errors
		for _, s := range []string{`"abc"`, `"123xyz"`, `"NaN"`, `"Infinity"`, `"-Infinity"`} {
			var h int8Holder
			assert.Error(t, json.Unmarshal([]byte(`{"val": `+s+`}`), &h, opts), "expected parse error for %s", s)
		}

		// Unsupported types (array, object)
		assert.Error(t, json.Unmarshal([]byte(`{"val": [123]}`), &dummyI8, opts))
		assert.Error(t, json.Unmarshal([]byte(`{"val": {"n": 123}}`), &dummyI8, opts))
	})

	t.Run("weaklyFloat", func(t *testing.T) {
		type floatHolder struct {
			F32 float32 `json:"f32"`
			F64 float64 `json:"f64"`
		}

		// Nulls
		var nulls floatHolder
		require.NoError(t, json.Unmarshal([]byte(`{"f32": null, "f64": null}`), &nulls, opts))
		assert.Equal(t, float32(0), nulls.F32)
		assert.Equal(t, float64(0), nulls.F64)

		// Empty strings
		var empties floatHolder
		require.NoError(t, json.Unmarshal([]byte(`{"f32": "", "f64": "  "}`), &empties, opts))
		assert.Equal(t, float32(0), empties.F32)
		assert.Equal(t, float64(0), empties.F64)

		// Numbers
		var nums floatHolder
		require.NoError(t, json.Unmarshal([]byte(`{"f32": 3.14, "f64": -2.71828}`), &nums, opts))
		assert.InDelta(t, float32(3.14), nums.F32, 0.001)
		assert.InDelta(t, float64(-2.71828), nums.F64, 0.00001)

		// Strings
		var strs floatHolder
		require.NoError(t, json.Unmarshal([]byte(`{"f32": "3.14", "f64": "-2.71828"}`), &strs, opts))
		assert.InDelta(t, float32(3.14), strs.F32, 0.001)
		assert.InDelta(t, float64(-2.71828), strs.F64, 0.00001)

		// Overflow for float32
		type f32Holder struct {
			Val float32 `json:"val"`
		}
		var dummyF32 f32Holder
		assert.Error(t, json.Unmarshal([]byte(`{"val": "1e40"}`), &dummyF32, opts))

		// Parse & Syntax Errors
		for _, s := range []string{`"abc"`, `"3.14.15"`, `"foo"`} {
			var h f32Holder
			assert.Error(t, json.Unmarshal([]byte(`{"val": `+s+`}`), &h, opts), "expected parse error for %s", s)
		}

		// Unsupported types (array, object)
		assert.Error(t, json.Unmarshal([]byte(`{"val": [3.14]}`), &dummyF32, opts))
		assert.Error(t, json.Unmarshal([]byte(`{"val": {"f": 3.14}}`), &dummyF32, opts))
	})
}

func TestDecode(t *testing.T) {
	opts := json.WithUnmarshalers(WeakUnmarshalers())

	input := map[string]any{
		"str":   100,
		"int":   "200",
		"uint":  300,
		"bool":  "yes",
		"float": "12.34",
	}

	var target weakTestStruct
	err := Decode(input, &target, opts)
	require.NoError(t, err)
	assert.Equal(t, "100", target.Str)
	assert.Equal(t, int64(200), target.Int)
	assert.Equal(t, uint32(300), target.Uint)
	assert.True(t, target.Bool)
	assert.Equal(t, 12.34, target.Float)
}

func BenchmarkWeakUnmarshalers(b *testing.B) {
	opts := json.WithUnmarshalers(WeakUnmarshalers())

	b.Run("Integer_Number", func(b *testing.B) {
		input := []byte(`{"i64": 123456789, "i32": 123, "u32": 456, "i8": 42}`)
		type S struct {
			I64 int64  `json:"i64"`
			I32 int32  `json:"i32"`
			U32 uint32 `json:"u32"`
			I8  int8   `json:"i8"`
		}
		var s S
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = json.Unmarshal(input, &s, opts)
		}
	})

	b.Run("Integer_FloatNumber", func(b *testing.B) {
		input := []byte(`{"i64": 12345.67, "i32": 123.4, "u32": 456.78, "i8": 42.1}`)
		type S struct {
			I64 int64  `json:"i64"`
			I32 int32  `json:"i32"`
			U32 uint32 `json:"u32"`
			I8  int8   `json:"i8"`
		}
		var s S
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = json.Unmarshal(input, &s, opts)
		}
	})

	b.Run("Integer_String", func(b *testing.B) {
		input := []byte(`{"i64": "123456789", "i32": "123", "u32": "456", "i8": "42"}`)
		type S struct {
			I64 int64  `json:"i64"`
			I32 int32  `json:"i32"`
			U32 uint32 `json:"u32"`
			I8  int8   `json:"i8"`
		}
		var s S
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = json.Unmarshal(input, &s, opts)
		}
	})

	b.Run("Integer_FloatString", func(b *testing.B) {
		input := []byte(`{"i64": "12345.67", "i32": "123.4", "u32": "456.78", "i8": "42.1"}`)
		type S struct {
			I64 int64  `json:"i64"`
			I32 int32  `json:"i32"`
			U32 uint32 `json:"u32"`
			I8  int8   `json:"i8"`
		}
		var s S
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = json.Unmarshal(input, &s, opts)
		}
	})

	b.Run("Float_Number", func(b *testing.B) {
		input := []byte(`{"f64": 12345.6789, "f32": 3.14}`)
		type S struct {
			F64 float64 `json:"f64"`
			F32 float32 `json:"f32"`
		}
		var s S
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = json.Unmarshal(input, &s, opts)
		}
	})

	b.Run("Float_String", func(b *testing.B) {
		input := []byte(`{"f64": "12345.6789", "f32": "3.14"}`)
		type S struct {
			F64 float64 `json:"f64"`
			F32 float32 `json:"f32"`
		}
		var s S
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = json.Unmarshal(input, &s, opts)
		}
	})
}

func BenchmarkDecode(b *testing.B) {
	opts := json.WithUnmarshalers(WeakUnmarshalers())
	input := map[string]any{
		"str":   100,
		"int":   "200",
		"uint":  300,
		"bool":  "yes",
		"float": "12.34",
	}
	var target weakTestStruct
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Decode(input, &target, opts)
	}
}
