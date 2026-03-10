/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package workflows

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
)

func (o *On) UnmarshalJSONFrom(d *jsontext.Decoder) error {
	switch k := d.PeekKind(); k {
	// 1. Shorthand for single event, e.g: `on: push`
	case jsontext.KindString:
		if tok, err := d.ReadToken(); err != nil {
			return err
		} else {
			e := tok.String()
			*o = map[string]Event{e: nil}
		}
		return nil

	// 2. Shorthand for multiple events, e.g: `on: [push, fork]`
	case jsontext.KindBeginArray:
		var events []string
		if err := json.UnmarshalDecode(d, &events); err != nil {
			return err
		}
		m := make(map[string]Event, len(events))
		for _, e := range events {
			m[e] = nil
		}
		*o = m
		return nil

	// 3. Full object format, e.g: `"on": {"label": {"types": ["created", "edited"]}}`
	case jsontext.KindBeginObject:
		return json.UnmarshalDecode(d, (*map[string]Event)(o))

	default:
		return fmt.Errorf("expected string, array or object for On, got kind %v", k)
	}
}
