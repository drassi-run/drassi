/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package model

import (
	"math"
	"reflect"
	"strconv"

	"github.com/go-viper/mapstructure/v2"
)

var typeListString = reflect.TypeFor[[]string]()
var typeMapString = reflect.TypeFor[map[string]any]()

func Stringify(data any) (string, bool) {
	v := reflect.ValueOf(data)
	switch k := v.Kind(); {
	case k == reflect.Invalid:
		return "", true
	case k == reflect.Bool:
		return strconv.FormatBool(v.Bool()), true
	case k <= reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), true
	case k <= reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), true
	case k <= reflect.Float64: // see Float.ToString()
		f := v.Float()
		if math.IsInf(f, 1) {
			return "Infinity", true
		} else if math.IsInf(f, -1) {
			return "-Infinity", true
		}
		return strconv.FormatFloat(f, 'G', 15, 64), true
	case k == reflect.String:
		return v.String(), true
	default:
		return "", false
	}
}

func ListStringify(data any) ([]string, bool) {
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Array && val.Kind() != reflect.Slice {
		return nil, false
	}

	// Use `Convert` to handle the underlying type
	if val.CanConvert(typeListString) {
		l := val.Convert(typeListString).Interface().([]string)
		return l, true
	}

	l := make([]string, val.Len())
	for i := 0; i < val.Len(); i++ {
		e := val.Index(i).Interface()
		s, ok := Stringify(e)
		if !ok {
			return nil, false
		}

		l[i] = s
	}
	return l, true
}

func ObjectStringify(data any) (map[string]any, bool) {
	val := reflect.ValueOf(data)
	switch k := val.Kind(); k {
	case reflect.Struct:
		return StructStringify(data)
	case reflect.Map:
		return MapStringify(data)
	default:
		return nil, false
	}
}

func MapStringify(data any) (map[string]any, bool) {
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Map {
		return nil, false
	}

	// Use `Convert` to handle the underlying type
	if val.CanConvert(typeMapString) {
		m := val.Convert(typeMapString).Interface().(map[string]any)
		return m, true
	}

	m := make(map[string]any, val.Len())
	iter := val.MapRange()
	for iter.Next() {
		k := iter.Key().Interface()
		s, ok := Stringify(k)
		if !ok {
			return nil, false
		}

		v := iter.Value().Interface()
		m[s] = v
	}
	return m, true
}

func StructStringify(data any) (map[string]any, bool) {
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Struct {
		return nil, false
	}

	m := make(map[string]any, val.NumField())
	if err := mapstructure.Decode(data, &m); err != nil {
		return nil, false
	}
	return m, true
}
