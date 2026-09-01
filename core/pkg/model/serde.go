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
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
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

func Decode(source any, target any, opts ...json.Options) (err error) {
	var data []byte
	switch s := source.(type) {
	case []byte:
		data = s
	case string:
		data = []byte(s)
	default:
		if data, err = json.Marshal(source); err != nil {
			return err
		}
	}
	return json.Unmarshal(data, target, opts...)
}

func WeakUnmarshalers() *json.Unmarshalers {
	return json.JoinUnmarshalers(weakUnmarshalers...)
}

var weakUnmarshalers []*json.Unmarshalers

func init() {
	weakUnmarshalers = []*json.Unmarshalers{
		json.UnmarshalFromFunc(weaklyString),
		json.UnmarshalFromFunc(weaklyBool),
		json.UnmarshalFromFunc(weaklyInteger[int64]()),
		json.UnmarshalFromFunc(weaklyInteger[int32]()),
		json.UnmarshalFromFunc(weaklyInteger[int16]()),
		json.UnmarshalFromFunc(weaklyInteger[int8]()),
		json.UnmarshalFromFunc(weaklyInteger[int]()),
		json.UnmarshalFromFunc(weaklyInteger[uint64]()),
		json.UnmarshalFromFunc(weaklyInteger[uint32]()),
		json.UnmarshalFromFunc(weaklyInteger[uint16]()),
		json.UnmarshalFromFunc(weaklyInteger[uint8]()),
		json.UnmarshalFromFunc(weaklyInteger[uint]()),
		json.UnmarshalFromFunc(weaklyFloat[float64]()),
		json.UnmarshalFromFunc(weaklyFloat[float32]()),
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

func weaklyBool(d *jsontext.Decoder, b *bool) error {
	switch k := d.PeekKind(); k {
	case jsontext.KindNull,
		jsontext.KindTrue, jsontext.KindFalse:
		if _, err := d.ReadToken(); err != nil {
			return err
		}
		*b = k == jsontext.KindTrue // null is false
	case jsontext.KindString:
		var s string
		if tok, err := d.ReadToken(); err != nil {
			return err
		} else {
			s = strings.TrimSpace(tok.String())
		}
		// all boolean values supported by YAML.
		// strconv.ParseBool is not support full list.
		switch strings.ToLower(s) {
		case "1", "yes", "on", "true":
			*b = true
		case "", "0", "no", "off", "false":
			*b = false
		default:
			return fmt.Errorf("invalid boolean string %q", s)
		}
	default:
		return errors.ErrUnsupported
	}
	return nil
}

func readScalarToken(d *jsontext.Decoder) (*jsontext.Token, error) {
	switch d.PeekKind() {
	case jsontext.KindNull:
		_, err := d.ReadToken()
		return nil, err
	case jsontext.KindNumber, jsontext.KindString:
		tok, err := d.ReadToken()
		return &tok, err
	default:
		return nil, errors.ErrUnsupported
	}
}

func castFloatToInt[I integer](f float64, t reflect.Type, signed bool) (I, error) {
	if math.IsNaN(f) {
		return 0, strconv.ErrRange
	}
	if signed {
		if f < math.MinInt64 || math.MaxInt64 < f || t.OverflowInt(int64(f)) {
			return 0, strconv.ErrRange
		}
		return I(int64(f)), nil
	}
	// unsigned
	if f < 0 || math.MaxUint64 < f || t.OverflowUint(uint64(f)) {
		return 0, strconv.ErrRange
	}
	return I(uint64(f)), nil
}

func parseInteger[I integer](tok *jsontext.Token, t reflect.Type) (I, error) {
	signed := t.Kind() <= reflect.Int64
	if tok.Kind() == jsontext.KindNumber {
		if signed {
			if v, err := tok.Int(); err == nil {
				if t.OverflowInt(v) {
					return 0, strconv.ErrRange
				}
				return I(v), nil
			} else if !errors.Is(err, strconv.ErrSyntax) {
				return 0, err
			}
		} else {
			if v, err := tok.Uint(); err == nil {
				if t.OverflowUint(v) {
					return 0, strconv.ErrRange
				}
				return I(v), nil
			} else if !errors.Is(err, strconv.ErrSyntax) {
				return 0, err
			}
		}
		if f, err := tok.Float(); err != nil {
			return 0, err
		} else {
			return castFloatToInt[I](f, t, signed)
		}
	}

	s := strings.TrimSpace(tok.String())
	if s == "" {
		return 0, nil
	}
	if signed {
		if v, err := strconv.ParseInt(s, 0, t.Bits()); err == nil {
			return I(v), nil
		} else if !errors.Is(err, strconv.ErrSyntax) {
			return 0, err
		}
	} else {
		if v, err := strconv.ParseUint(s, 0, t.Bits()); err == nil {
			return I(v), nil
		} else if !errors.Is(err, strconv.ErrSyntax) {
			return 0, err
		}
	}
	if f, err := strconv.ParseFloat(s, 64); err != nil {
		return 0, err
	} else {
		return castFloatToInt[I](f, t, signed)
	}
}

type integer interface {
	~int64 | ~int32 | ~int16 | ~int8 | ~int | ~uint64 | ~uint32 | ~uint16 | ~uint8 | ~uint
}

func weaklyInteger[I integer]() func(d *jsontext.Decoder, n *I) error {
	t := reflect.TypeFor[I]()

	return func(d *jsontext.Decoder, n *I) error {
		tok, err := readScalarToken(d)
		if err != nil || tok == nil {
			*n = 0
			return err
		}
		if v, err := parseInteger[I](tok, t); err != nil {
			return err
		} else {
			*n = v
			return nil
		}
	}
}

type float interface {
	~float64 | ~float32
}

func weaklyFloat[F float]() func(d *jsontext.Decoder, f *F) error {
	t := reflect.TypeFor[F]()

	return func(d *jsontext.Decoder, f *F) error {
		tok, err := readScalarToken(d)
		if err != nil || tok == nil {
			*f = 0
			return err
		}
		var v float64
		if tok.Kind() == jsontext.KindNumber {
			v, err = tok.Float()
		} else if s := strings.TrimSpace(tok.String()); s != "" {
			v, err = strconv.ParseFloat(s, t.Bits())
		}
		if err != nil {
			return err
		}
		if t.OverflowFloat(v) {
			return strconv.ErrRange
		}
		*f = F(v)
		return nil
	}
}
