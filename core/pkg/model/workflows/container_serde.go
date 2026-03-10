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

func (c *Container) UnmarshalJSONFrom(d *jsontext.Decoder) error {
	switch k := d.PeekKind(); k {
	// 1. Shorthand string format (e.g., "ubuntu:22.04")
	case jsontext.KindString:
		return json.UnmarshalDecode(d, &c.Image)

	// 2. Standard object format (e.g., {"image": "ubuntu:22.04", "env": {...}})
	case jsontext.KindBeginObject:
		type alias Container
		return json.UnmarshalDecode(d, (*alias)(c))

	default:
		return fmt.Errorf("expected string or object for Container, got kind %v", k)
	}
}
