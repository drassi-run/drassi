/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package model

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
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
