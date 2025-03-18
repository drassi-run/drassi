/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package workflows

import (
	"reflect"
	"strings"

	"drassi.run/core/pkg/model"
)

var typeToken = reflect.TypeFor[Token]()

func DecodeTokenHook(from reflect.Value, to reflect.Value) (any, error) {
	if !from.IsValid() || !to.Type().Implements(typeToken) || to.Interface() != nil {
		return valueOf(from), nil
	}

	var (
		token Token = nil
		data  any   = nil
	)
	switch from.Kind() {
	case reflect.Bool:
		token = NewLiteralToken(from.Bool())
	case reflect.String:
		s := from.String()
		if strings.Contains(s, OpenExpression) {
			token = NewExpressionToken(s)
		} else {
			token = NewLiteralToken(s)
		}
	case reflect.Slice, reflect.Array:
		token = NewSequenceToken(nil)
		data = from.Interface()
	case reflect.Map:
		token = NewMappingToken(nil)
		data = from.Interface()
	default:
		if from.CanInt() {
			token = NewLiteralToken(from.Int())
		} else if from.CanUint() {
			token = NewLiteralToken(from.Uint())
		} else if from.CanFloat() {
			token = NewLiteralToken(from.Float())
		} else {
			data = from.Interface()
		}
	}
	if token != nil {
		to.Set(reflect.ValueOf(token))
	}
	if data != nil {
		return data, nil
	} else {
		return token, nil
	}
}

func (m mappingToken) DecodeMapstructure(input any) (any, error) {
	inputVal := reflect.ValueOf(input)
	if inputVal.Kind() != reflect.Map {
		return input, nil
	}

	a := make([][2]any, 0, inputVal.Len())
	mapIter := inputVal.MapRange()
	for mapIter.Next() {
		key := mapIter.Key()
		val := mapIter.Value()

		pair := [2]any{key.Interface(), val.Interface()}
		a = append(a, pair)
	}
	return a, nil
}

func init() {
	model.RegisterDecodeHook(DecodeTokenHook)
}
