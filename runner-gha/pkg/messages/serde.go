/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package messages

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"time"
)

var zeroTime time.Time
var unmarshalers []*json.Unmarshalers

func init() {
	zeroTime, _ = time.Parse(time.TimeOnly, "00:00:00")

	unmarshalers = append(unmarshalers,
		json.UnmarshalFromFunc(unmarshalDuration),
		json.UnmarshalFromFunc(unmarshalTime),
		json.UnmarshalFromFunc(unmarshalValue),
	)
}

func JsonOptions() []json.Options {
	um := json.JoinUnmarshalers(unmarshalers...)
	return []json.Options{
		json.WithUnmarshalers(um),
	}
}

func Decode[M any](content []byte) (*M, error) {
	m := new(M)
	if err := json.Unmarshal(content, &m, JsonOptions()...); err != nil {
		return nil, err
	}
	return m, nil
}

func unmarshalDuration(dec *jsontext.Decoder, val *time.Duration) error {
	if dec.PeekKind() != jsontext.KindString {
		return errors.New("expected string")
	}

	tok, err := dec.ReadToken() // Consume the string token
	if err != nil {
		return err
	}

	s := tok.String()
	if d := parseDuration(s); d != nil {
		*val = *d
		return nil
	}
	return errors.New("unknown duration format: " + s)
}

func parseDuration(s string) *time.Duration {
	// first try (go time.Duration format)
	if d, err := time.ParseDuration(s); err == nil {
		return &d
	}

	// second try (C# Timespan format)
	if t, err := time.Parse(time.TimeOnly, s); err == nil {
		return new(t.Sub(zeroTime))
	}

	return nil
}

func unmarshalTime(dec *jsontext.Decoder, val *time.Time) error {
	if dec.PeekKind() != jsontext.KindString {
		return errors.New("expected string")
	}

	tok, err := dec.ReadToken() // Consume the string token
	if err != nil {
		return err
	}

	s := tok.String()
	t, err := parseTime(s)
	if err != nil {
		return err
	}

	*val = t
	return nil
}

var timeFormats = []string{
	time.RFC3339,          // RFC3339 format with zone info
	"2006-01-02T15:04:05", // RFC3339 format w/o zone info
}

func parseTime(s string) (time.Time, error) {
	for _, format := range timeFormats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, errors.New("unknown time format: " + s)
}

func (t *TemplateToken) UnmarshalJSONFrom(d *jsontext.Decoder) error {
	switch k := d.PeekKind(); k {
	// 1. Shorthand string format (e.g., "script")
	case jsontext.KindString:
		t.Type = TokenTypeString
		return json.UnmarshalDecode(d, &t.String)

	// 2. Standard object format (e.g., {"type": 5, "bool": true})
	case jsontext.KindBeginObject:
		type alias TemplateToken
		return json.UnmarshalDecode(d, (*alias)(t))

	default:
		return fmt.Errorf("expected string or object for TemplateToken, got kind %v", k)
	}
}

func unmarshalValue(dec *jsontext.Decoder, val *Value) error {
	var a any

	switch k := dec.PeekKind(); k {
	case jsontext.KindBeginObject:
		return unmarshalMapValue(dec, val)
	case jsontext.KindBeginArray:
		a = make([]Value, 0)
	}

	if err := json.UnmarshalDecode(dec, &a); err != nil {
		return err
	}
	*val = Value(a)
	return nil
}

func unmarshalMapValue(dec *jsontext.Decoder, val *Value) error {
	type entry struct {
		Key   string `json:"k,omitempty"`
		Value Value  `json:"v,omitempty"`
	}
	// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/ContextData/PipelineContextData.cs
	// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/ContextData/PipelineContextDataType.cs
	type value struct {
		Type       int     `json:"t"`
		String     string  `json:"s,omitempty"` // Type=0 (StringContextData.cs)
		Array      []Value `json:"a,omitempty"` // Type=1 (ArrayContextData.cs)
		Dictionary []entry `json:"d,omitempty"` // Type=2 (DictionaryContextData.cs)
		Boolean    bool    `json:"b,omitempty"` // Type=3 (BooleanContextData.cs)
		Number     float64 `json:"n,omitempty"` // Type=4 (NumberContextData.cs)
	}
	var v value

	if err := json.UnmarshalDecode(dec, &v); err != nil {
		return err
	}

	switch v.Type {
	case 0:
		*val = v.String
	case 1:
		*val = v.Array
	case 2:
		m := make(map[string]Value, len(v.Dictionary))
		for _, d := range v.Dictionary {
			m[d.Key] = d.Value
		}
		*val = m
	case 3:
		*val = v.Boolean
	case 4:
		*val = v.Number
	case 5: // case-sensitive dictionary
		// TODO: *val = v.Dictionary
	default:
		return fmt.Errorf("unknown Value type=%d", v.Type)
	}
	return nil
}
