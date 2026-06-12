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
	"regexp"
	"strconv"
	"strings"
	"time"
)

var unmarshalers []*json.Unmarshalers

func init() {
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
	d, err := parseDuration(s)
	if err != nil {
		return err
	}

	*val = d
	return nil
}

func parseDuration(s string) (time.Duration, error) {
	// first try (go time.Duration format)
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}

	// second try (C# Timespan format)
	return parseTimespan(s)
}

// Regex to match C# TimeSpan format: [-][d.]hh:mm:ss[.fffffff]
var timeSpanRegex = regexp.MustCompile(`^(-)?(?:(\d+)\.)?(\d{1,2}):(\d{1,2}):(\d{1,2})(?:\.(\d+))?$`)

func parseTimespan(s string) (time.Duration, error) {
	matches := timeSpanRegex.FindStringSubmatch(s)
	if matches == nil {
		return 0, errors.New("unknown TimeSpan format: " + s)
	}
	isNegative := matches[1] == "-"
	days := parseInt(matches[2])
	hours := parseInt(matches[3])
	minutes := parseInt(matches[4])
	seconds := parseInt(matches[5])
	fraction := matches[6]

	var nanoseconds int64
	if fraction != "" {
		// Go time.Duration uses nanoseconds (9 decimal places).
		// We must pad the C# fraction (which is up to 7 places) with zeros
		// so that "1" (0.1s) becomes "100000000" ns.
		if len(fraction) > 9 {
			fraction = fraction[:9] // Truncate sub-nanosecond precision if it somehow exists
		} else if len(fraction) < 9 {
			fraction += strings.Repeat("0", 9-len(fraction))
		}
		nanoseconds = parseInt(fraction)
	}
	if hours >= 24 {
		return 0, fmt.Errorf("hours out of range (%d)", hours)
	}
	if minutes >= 60 {
		return 0, fmt.Errorf("minutes out of range (%d)", minutes)
	}
	if seconds >= 60 {
		return 0, fmt.Errorf("seconds out of range (%d)", seconds)
	}

	d := time.Duration(nanoseconds)
	d += time.Duration(seconds) * time.Second
	d += time.Duration(minutes) * time.Minute
	d += time.Duration(hours) * time.Hour
	d += time.Duration(days*24) * time.Hour

	if isNegative {
		d = -d
	}
	return d, nil
}

func parseInt(s string) int64 {
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
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
