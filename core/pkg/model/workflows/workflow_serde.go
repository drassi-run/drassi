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

func (p *Permissions) UnmarshalJSONFrom(d *jsontext.Decoder) error {
	switch k := d.PeekKind(); k {
	// 1. Shorthand string format: read-all | write-all
	case jsontext.KindString:
		tok, err := d.ReadToken()
		if err != nil {
			return err
		}
		switch l := tok.String(); l {
		case "read-all":
			*p = Permissions{"*": PermissionsLevelRead}
		case "write-all":
			*p = Permissions{"*": PermissionsLevelWrite}
		default:
			return fmt.Errorf("unknown permission %s", l)
		}
		return nil

	// 2. Standard object format (e.g., {"actions": "read"})
	case jsontext.KindBeginObject:
		return json.UnmarshalDecode(d, (*map[string]PermissionsLevel)(p))

	default:
		return fmt.Errorf("expected string or object for Permission, got kind %v", k)
	}
}

func (c *Concurrency) UnmarshalJSONFrom(d *jsontext.Decoder) error {
	switch k := d.PeekKind(); k {
	// 1. Shorthand string format (e.g., "group1")
	case jsontext.KindString:
		return json.UnmarshalDecode(d, &c.Group)

	// 2. Standard object format (e.g., {"group": "group1", "cancel-in-progress": true})
	case jsontext.KindBeginObject:
		type alias Concurrency
		return json.UnmarshalDecode(d, (*alias)(c))

	default:
		return fmt.Errorf("expected string or object for Environment, got kind %v", k)
	}
}
