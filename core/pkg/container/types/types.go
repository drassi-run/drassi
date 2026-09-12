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
	"strings"

	"github.com/docker/go-units"
)

// Mapping is a mapping type that can be converted from a list of key[=value] strings.
// For the key with an empty value (`key=`), or key without value (`key`), the
// mapped value is set to an empty string `""`.
//   - [github.com/compose-spec/compose-go/v2/types.Mapping]
type Mapping map[string]string

func (m *Mapping) UnmarshalJSONFrom(d *jsontext.Decoder) error {
	switch kind := d.PeekKind(); kind {
	case jsontext.KindBeginObject:
		type alias Mapping
		return json.UnmarshalDecode(d, (*alias)(m))
	case jsontext.KindBeginArray:
		var list []string
		if err := json.UnmarshalDecode(d, &list); err != nil {
			return err
		}
		res := make(Mapping, len(list))
		for _, item := range list {
			k, v, _ := strings.Cut(item, "=")
			res[k] = v
		}
		*m = res
		return nil
	case jsontext.KindNull:
		if _, err := d.ReadToken(); err != nil {
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

func (u *UnitBytes) UnmarshalJSONFrom(d *jsontext.Decoder) error {
	switch kind := d.PeekKind(); kind {
	case jsontext.KindNumber:
		return json.UnmarshalDecode(d, (*int64)(u))
	case jsontext.KindString:
		var s string
		if err := json.UnmarshalDecode(d, &s); err != nil {
			return err
		}
		if b, err := units.RAMInBytes(s); err != nil {
			return fmt.Errorf("invalid byte size %q: %w", s, err)
		} else {
			*u = UnitBytes(b)
			return nil
		}
	case jsontext.KindNull:
		if _, err := d.ReadToken(); err != nil {
			return err
		}
		*u = 0
		return nil
	default:
		return fmt.Errorf("expected number, string, or null for UnitBytes, got %v", kind)
	}
}
