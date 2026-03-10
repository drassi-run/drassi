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

	"drassi.run/core/pkg/model"
)

var zeroTime time.Time

func init() {
	zeroTime, _ = time.Parse(time.TimeOnly, "00:00:00")
}

func Decode[M any](content []byte) (*M, error) {
	var a any
	if err := json.Unmarshal(content, &a); err != nil {
		return nil, err
	}

	m := new(M)
	if err := model.Decode(a, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (d *Duration) DecodeMapstructure(input any) (any, error) {
	if s, ok := input.(string); ok {
		return nil, d.parse(s)
	} else {
		return input, nil
	}
}

func (d *Duration) parse(s string) error {
	// first try (go time.Duration format)
	if duration, err := time.ParseDuration(s); err == nil {
		d.Duration = duration
		return nil
	}

	// second try (C# Timespan format)
	if t, err := time.Parse(time.TimeOnly, s); err == nil {
		d.Duration = t.Sub(zeroTime)
		return nil
	}

	return errors.New("unknown duration format: " + s)
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

func (t *Time) DecodeMapstructure(input any) (any, error) {
	if s, ok := input.(string); ok {
		return nil, t.parse(s)
	} else {
		return input, nil
	}
}

func (t *Time) parse(s string) error {
	// first try (RFC3339 format with zone info)
	if dt, err := time.Parse(time.RFC3339, s); err == nil {
		t.Time = dt
		return nil
	}

	// second try (RFC3339 format w/o zone info)
	if dt, err := time.Parse("2006-01-02T15:04:05", s); err == nil {
		t.Time = dt
		return nil
	}

	return errors.New("unknown time format: " + s)
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
	if t := parseTime(s); t != nil {
		*val = *t
		return nil
	}
	return errors.New("unknown time format: " + s)
}

func parseTime(s string) *time.Time {
	// first try (RFC3339 format with zone info)
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return &t
	}

	// second try (RFC3339 format w/o zone info)
	if t, err := time.Parse("2006-01-02T15:04:05", s); err == nil {
		return &t
	}

	return nil
}

func (t *TemplateToken) DecodeMapstructure(input any) (any, error) {
	if s, ok := input.(string); ok {
		t.Type = TokenTypeString
		t.String = s
		return nil, nil
	} else {
		return input, nil
	}
}

func (t *TemplateToken) UnmarshalJSONFrom(d *jsontext.Decoder) error {
	switch k := d.PeekKind(); k {
	// 1. Shorthand string format (e.g., "foobar")
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

func (cd *ContextData) DecodeMapstructure(input any) (any, error) {
	m, ok := input.(map[string]any)
	if !ok {
		return input, nil
	}

	var err error
	var res = make(map[string]any, len(m))

	for k, v := range m {
		if res[k], err = DecodeContextData(k, v); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func DecodeContextData(name string, d any) (any, error) {
	if d == nil {
		return d, nil
	}
	switch v := d.(type) {
	case map[string]any:
		return decodeGenericContextData(name, v)
	case []any:
		return decodeArrayContextData(name, v)
	case string, bool:
		return v, nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	default:
		return nil, fmt.Errorf("contextData %s: unsupported data type, %T", name, d)
	}
}

func decodeGenericContextData(name string, o map[string]any) (any, error) {
	t, ok := o["t"].(float64)
	if !ok {
		return nil, fmt.Errorf("ContextData %s required field 't'", name)
	}
	typ := int(t)

	// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/ContextData/PipelineContextData.cs
	// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/ContextData/PipelineContextDataType.cs
	switch typ {
	// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/ContextData/StringContextData.cs
	case 0: // string
		if s, ok := o["s"]; !ok {
			return "", nil // default value
		} else if str, ok := s.(string); ok {
			return str, nil
		} else {
			return nil, fmt.Errorf("field 's' in StringContextData %s must be a string, got %T", name, s)
		}

	//	https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/ContextData/ArrayContextData.cs
	case 1: // array
		if a, ok := o["a"]; !ok {
			return nil, nil // default value
		} else if array, ok := a.([]any); ok {
			return decodeArrayContextData(name, array)
		} else {
			return nil, fmt.Errorf("field 'a' in ArrayContextData %s must be a array, got %T", name, a)
		}

	//	https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/ContextData/DictionaryContextData.cs
	case 2: // dictionary
		if d, ok := o["d"]; !ok {
			return nil, nil // default value
		} else if dic, ok := d.([]map[string]any); ok {
			return decodeMapContextData(name, dic)
		} else if dic, ok := d.([]any); ok {
			return decodeMapContextData(name, dic)
		} else {
			return nil, fmt.Errorf("field 'd' in DictionaryContextData %s must be a map, got %T", name, d)
		}

	// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/ContextData/BooleanContextData.cs
	case 3: // boolean
		if b, ok := o["b"]; !ok {
			return false, nil // default value
		} else if boolean, ok := b.(bool); ok {
			return boolean, nil
		} else {
			return nil, fmt.Errorf("field 'b' in BooleanContextData %s must be a boolean, got %T", name, b)
		}

	//	https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTPipelines/Pipelines/ContextData/NumberContextData.cs
	case 4: // number
		if n, ok := o["n"]; !ok {
			return 0, nil // default value
		} else if number, ok := n.(float64); ok {
			return number, nil
		} else {
			return nil, fmt.Errorf("field 'n' in NumberContextData %s must be a number, got %T", name, n)
		}

	case 5: // case-sensitive dictionary
		return nil, nil // TODO

	default:
		return nil, fmt.Errorf("ContextData %s unknown type %d", name, typ)
	}
}

func decodeMapContextData[E any](name string, m []E) (map[string]any, error) {
	o := make(map[string]any, len(m))
	for i, pair := range m {
		p, ok := any(pair).(map[string]any)
		if !ok {
			return nil, fmt.Errorf("ContextData %s item #%d has invalid type %T", name, i, pair)
		}
		k, ok := p["k"].(string)
		if !ok {
			return nil, fmt.Errorf("ContextData %s item #%d require field 'k' of string", name, i)
		}
		v, ok := p["v"]
		if !ok {
			return nil, fmt.Errorf("ContextData %s item #%d require field 'v'", name, i)
		}
		n := fmt.Sprintf("%s.%s", name, k)
		v, err := DecodeContextData(n, v)
		if err != nil {
			return nil, err
		}
		o[k] = v
	}
	return o, nil
}

func decodeArrayContextData[E any](name string, a []E) ([]any, error) {
	o := make([]any, len(a))
	var err error
	for i, v := range a {
		n := fmt.Sprintf("%s[%d]", name, i)
		if o[i], err = DecodeContextData(n, v); err != nil {
			return nil, err
		}
	}
	return o, nil
}
