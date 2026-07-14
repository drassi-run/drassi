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
	"strings"
)

func init() {
	u := json.UnmarshalFromFunc(unmarshalToken)
	unmarshalers = append(unmarshalers, u)
}

func unmarshalToken(d *jsontext.Decoder, t *Token) error {
	switch d.PeekKind() {
	case jsontext.KindNull:
		if _, err := d.ReadToken(); err != nil {
			return err
		}
		*t = nil

	case jsontext.KindTrue, jsontext.KindFalse:
		if tok, err := d.ReadToken(); err != nil {
			return err
		} else {
			b := tok.Bool()
			*t = NewLiteralToken(b)
		}

	case jsontext.KindNumber:
		if tok, err := d.ReadToken(); err != nil {
			return err
		} else {
			f := tok.Float()
			*t = NewLiteralToken(f)
		}

	case jsontext.KindString:
		if tok, err := d.ReadToken(); err != nil {
			return err
		} else {
			s := tok.String()
			if strings.Contains(s, OpenExpression) {
				*t = NewExpressionToken(s)
			} else {
				*t = NewLiteralToken(s)
			}
		}

	case jsontext.KindBeginArray:
		var seq = make(sequenceToken, 0)
		if err := json.UnmarshalDecode(d, &seq); err != nil {
			return err
		}
		*t = seq

	case jsontext.KindBeginObject:
		var dic = make(mappingToken, 0)
		if err := dic.unmarshalJsonMap(d); err != nil {
			return err
		}
		*t = dic

	default:
		return fmt.Errorf("unknown token type %v", d.PeekKind())
	}
	return nil
}

func (m *mappingToken) UnmarshalJSONFrom(d *jsontext.Decoder) error {
	switch k := d.PeekKind(); k {
	case jsontext.KindBeginObject:
		*m = make(mappingToken, 0)
		return m.unmarshalJsonMap(d)
	default:
		return fmt.Errorf("unknown token type %v", k)
	}
}

func (m *mappingToken) unmarshalJsonMap(d *jsontext.Decoder) error {
	// Consume "{"
	if _, err := d.ReadToken(); err != nil {
		return err
	}

	for d.PeekKind() != jsontext.KindEndObject {
		var pair [2]Token

		// Object key
		if err := json.UnmarshalDecode(d, &pair[0]); err != nil {
			return err
		}

		// Object value.
		if err := json.UnmarshalDecode(d, &pair[1]); err != nil {
			return err
		}

		*m = append(*m, pair)
	}

	// Consume "}".
	if _, err := d.ReadToken(); err != nil {
		return err
	}
	return nil
}
