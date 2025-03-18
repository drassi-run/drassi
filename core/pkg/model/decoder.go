/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package model

import "reflect"

// comparable to yaml.Unmarshaler, decoder allow a type to define its own custom logic to convert value
// see https://github.com/go-viper/mapstructure/v2/pull/294
// see https://github.com/mitchellh/mapstructure/pull/294
type decoder interface {
	DecodeMapstructure(any) (any, error)
}

// see https://github.com/go-viper/mapstructure/v2/issues/115#issuecomment-735287466
// see https://github.com/mitchellh/mapstructure/issues/115#issuecomment-735287466
// adapted to support types derived from built-in types, as DecodeMapstructure would not be able to mutate internal
// value, so need to invoke DecodeMapstructure defined by pointer to type
func decoderHook(from reflect.Value, to reflect.Value) (any, error) {
	// If the destination implements the decoder interface
	t, ok := to.Interface().(decoder)
	if !ok {
		// for non-struct types, we need to invoke func (*type) DecodeMapstructure()
		if to.CanAddr() {
			pto := to.Addr()
			t, ok = pto.Interface().(decoder)
		}
		if !ok {
			return from.Interface(), nil
		}
	}
	// If the destination is a nil pointer, create and assign the target value first
	if typ := to.Type(); typ.Kind() == reflect.Pointer && to.IsNil() {
		to.Set(reflect.New(typ.Elem()))
		t = to.Interface().(decoder)
	}

	// Call the custom DecodeMapstructure method
	f := from.Interface()
	if d, err := t.DecodeMapstructure(f); err != nil {
		return nil, err
	} else if d != nil {
		return d, nil
	} else {
		// d == nil: all inputs already processed
		return t, nil
	}
}
