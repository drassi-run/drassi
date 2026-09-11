/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"strconv"
	"strings"

	"github.com/docker/go-units"
)

// Mapping is a mapping type that can be unmarshaled from either a JSON object
// or a list of "KEY=VALUE" (or "KEY") strings, matching Docker Compose syntax.
type Mapping map[string]string

var _ json.UnmarshalerFrom = (*Mapping)(nil)

func (m *Mapping) UnmarshalJSONFrom(d *jsontext.Decoder) error {
	switch kind := d.PeekKind(); kind {
	case jsontext.KindBeginObject:
		var raw map[string]any
		if err := json.UnmarshalDecode(d, &raw); err != nil {
			return err
		}
		res := make(Mapping, len(raw))
		for k, v := range raw {
			if v == nil {
				res[k] = ""
			} else {
				res[k] = fmt.Sprint(v)
			}
		}
		*m = res
		return nil
	case jsontext.KindBeginArray:
		var list []any
		if err := json.UnmarshalDecode(d, &list); err != nil {
			return err
		}
		res := make(Mapping, len(list))
		for _, item := range list {
			str := fmt.Sprint(item)
			k, v, ok := strings.Cut(str, "=")
			if ok {
				res[k] = v
			} else {
				res[k] = ""
			}
		}
		*m = res
		return nil
	case jsontext.KindNull:
		var v any
		if err := json.UnmarshalDecode(d, &v); err != nil {
			return err
		}
		*m = nil
		return nil
	default:
		return fmt.Errorf("expected object, array, or null for Mapping, got %v", kind)
	}
}

// UnitBytes represents byte sizes, supporting unmarshaling from both
// integer numbers and human-readable unit strings (e.g. "2g", "512m", "64mb").
type UnitBytes int64

var _ json.UnmarshalerFrom = (*UnitBytes)(nil)

func (u *UnitBytes) parseString(s string) error {
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		*u = UnitBytes(n)
		return nil
	}
	b, err := units.RAMInBytes(s)
	if err != nil {
		return fmt.Errorf("invalid byte size %q: %w", s, err)
	}
	*u = UnitBytes(b)
	return nil
}

func (u *UnitBytes) UnmarshalJSONFrom(d *jsontext.Decoder) error {
	switch kind := d.PeekKind(); kind {
	case jsontext.KindNumber:
		var n int64
		if err := json.UnmarshalDecode(d, &n); err != nil {
			return err
		}
		*u = UnitBytes(n)
		return nil
	case jsontext.KindString:
		var s string
		if err := json.UnmarshalDecode(d, &s); err != nil {
			return err
		}
		return u.parseString(s)
	case jsontext.KindNull:
		var v any
		if err := json.UnmarshalDecode(d, &v); err != nil {
			return err
		}
		*u = 0
		return nil
	default:
		return fmt.Errorf("expected number, string, or null for UnitBytes, got %v", kind)
	}
}
