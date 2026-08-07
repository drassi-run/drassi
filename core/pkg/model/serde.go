/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package model

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"reflect"
)

func UnmarshalInterface[T any](dis func(raw jsontext.Value) (T, error)) *json.Unmarshalers {
	fn := func(d *jsontext.Decoder, val *T) error {
		if d.PeekKind() == jsontext.KindNull {
			if _, err := d.ReadToken(); err != nil {
				return err
			}
			var zero T
			*val = zero
			return nil
		}

		var raw jsontext.Value
		if err := json.UnmarshalDecode(d, &raw); err != nil {
			return err
		}

		if t, err := dis(raw); err != nil {
			return err
		} else if err = json.Unmarshal(raw, t, d.Options()); err != nil {
			return err
		} else {
			*val = t
			return nil
		}
	}
	return json.UnmarshalFromFunc(fn)
}

func Decode(source any, target any, opts ...json.Options) error {
	var data []byte
	switch s := source.(type) {
	case []byte:
		data = s
	case string:
		data = []byte(s)
	default:
		var err error
		data, err = json.Marshal(source)
		if err != nil {
			return err
		}
	}
	options := []json.Options{
		json.WithUnmarshalers(WeakUnmarshalers()),
	}
	return json.Unmarshal(data, target, append(options, opts...)...)
}

func WeakUnmarshalers() *json.Unmarshalers {
	return json.JoinUnmarshalers(weakUnmarshalers...)
}

var weakUnmarshalers []*json.Unmarshalers

func init() {
	weakUnmarshalers = []*json.Unmarshalers{
		json.UnmarshalFromFunc(weaklyString),
		json.UnmarshalFromFunc(weaklyNumber[int64]()),
		json.UnmarshalFromFunc(weaklyNumber[int32]()),
		json.UnmarshalFromFunc(weaklyNumber[int16]()),
		json.UnmarshalFromFunc(weaklyNumber[int8]()),
		json.UnmarshalFromFunc(weaklyNumber[int]()),
		json.UnmarshalFromFunc(weaklyNumber[uint64]()),
		json.UnmarshalFromFunc(weaklyNumber[uint32]()),
		json.UnmarshalFromFunc(weaklyNumber[uint16]()),
		json.UnmarshalFromFunc(weaklyNumber[uint8]()),
		json.UnmarshalFromFunc(weaklyNumber[uint]()),
	}
}

func weaklyString(d *jsontext.Decoder, s *string) error {
	switch k := d.PeekKind(); k {
	case jsontext.KindNull:
		if _, err := d.ReadToken(); err != nil {
			return err
		}
		*s = ""
		return nil
	case jsontext.KindTrue, jsontext.KindFalse,
		jsontext.KindString, jsontext.KindNumber:
		if tok, err := d.ReadToken(); err != nil {
			return err
		} else {
			*s = tok.String()
			return nil
		}
	default:
		return errors.ErrUnsupported
	}
}

type integer interface {
	~int64 | ~int32 | ~int16 | ~int8 | ~int | ~uint64 | ~uint32 | ~uint16 | ~uint8 | ~uint
}

func weaklyNumber[I integer]() func(d *jsontext.Decoder, n *I) error {
	t := reflect.TypeFor[I]()
	signed := t.Kind() <= reflect.Int64

	return func(d *jsontext.Decoder, n *I) error {
		if d.PeekKind() != jsontext.KindNumber {
			return errors.ErrUnsupported
		}

		tok, err := d.ReadToken()
		if err != nil {
			return err
		}

		if signed {
			if v, err := tok.Int(); err == nil {
				*n = I(v)
				return nil
			}
			f, err := tok.Float()
			if err != nil {
				return err
			}
			*n = I(int64(f))
		} else {
			if v, err := tok.Uint(); err == nil {
				*n = I(v)
				return nil
			}
			f, err := tok.Float()
			if err != nil {
				return err
			}
			*n = I(uint64(f))
		}

		return nil
	}
}
