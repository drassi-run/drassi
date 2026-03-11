/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package messages

import (
	"encoding/json/v2"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDuration(t *testing.T) {
	t.Run("go_duration_format", func(t *testing.T) {
		cases := map[string]time.Duration{
			"0s":     0,
			"1h30m":  1*time.Hour + 30*time.Minute,
			"300ms":  300 * time.Millisecond,
			"2h3m4s": 2*time.Hour + 3*time.Minute + 4*time.Second,
		}
		for i, d := range cases {
			t.Run(i, func(t *testing.T) {
				got := parseDuration(i)
				assert.NotNil(t, got)
				assert.Equal(t, d, *got)
			})
		}
	})

	t.Run("csharp_timespan_format", func(t *testing.T) {
		cases := map[string]time.Duration{
			"00:00:00": 0,
			"01:30:00": 1*time.Hour + 30*time.Minute,
			"00:00:05": 5 * time.Second,
			"02:03:04": 2*time.Hour + 3*time.Minute + 4*time.Second,
		}
		for i, d := range cases {
			t.Run(i, func(t *testing.T) {
				got := parseDuration(i)
				assert.NotNil(t, got)
				assert.Equal(t, d, *got)
			})
		}
	})

	t.Run("unknown_format", func(t *testing.T) {
		cases := []string{
			"",
			"not-a-duration",
			"1234",
			"P1DT2H", // ISO 8601 – not supported
		}
		for _, tc := range cases {
			t.Run(tc, func(t *testing.T) {
				got := parseDuration(tc)
				assert.Nil(t, got)
			})
		}
	})
}

func TestUnmarshalDuration(t *testing.T) {
	unmarshalers := json.UnmarshalFromFunc(unmarshalDuration)

	t.Run("valid", func(t *testing.T) {
		cases := map[string]time.Duration{
			"1h30m":    1*time.Hour + 30*time.Minute, // go duration format
			"01:30:00": 1*time.Hour + 30*time.Minute, // C# timespan format
		}
		for i, d := range cases {
			t.Run(i, func(t *testing.T) {
				var got time.Duration
				in := fmt.Appendf(nil, "%q", i)
				err := json.Unmarshal(in, &got, json.WithUnmarshalers(unmarshalers))
				assert.NoError(t, err)
				assert.Equal(t, d, got)
			})
		}
	})

	t.Run("invalid", func(t *testing.T) {
		cases := []string{
			"3600",             // non string token
			`"not-a-duration"`, // invalid string
		}
		for _, s := range cases {
			t.Run(s, func(t *testing.T) {
				var got time.Duration
				err := json.Unmarshal([]byte(s), &got, json.WithUnmarshalers(unmarshalers))
				assert.Error(t, err)
			})
		}
	})
}

func TestParseTime(t *testing.T) {
	t.Run("rfc3339_format", func(t *testing.T) {
		cases := map[string]time.Time{
			"2024-03-11T10:00:00Z":      time.Date(2024, 3, 11, 10, 0, 0, 0, time.UTC),
			"2024-03-11T10:00:00+07:00": time.Date(2024, 3, 11, 10, 0, 0, 0, time.FixedZone("", 7*3600)),
		}
		for i, v := range cases {
			t.Run(i, func(t *testing.T) {
				got := parseTime(i)
				assert.NotNil(t, got)
				assert.True(t, v.Equal(*got))
			})
		}
	})

	t.Run("no_zone_format", func(t *testing.T) {
		cases := map[string]time.Time{
			"2024-03-11T10:00:00": time.Date(2024, 3, 11, 10, 0, 0, 0, time.UTC),
		}
		for i, v := range cases {
			t.Run(i, func(t *testing.T) {
				got := parseTime(i)
				assert.NotNil(t, got)
				assert.True(t, v.Equal(*got))
			})
		}
	})

	t.Run("unknown_format", func(t *testing.T) {
		cases := []string{
			"",
			"not-a-time",
			"2024-03-11",
		}
		for _, tc := range cases {
			t.Run(tc, func(t *testing.T) {
				got := parseTime(tc)
				assert.Nil(t, got)
			})
		}
	})
}

func TestUnmarshalTime(t *testing.T) {
	unmarshalers := json.UnmarshalFromFunc(unmarshalTime)

	t.Run("valid", func(t *testing.T) {
		cases := map[string]time.Time{
			"2024-03-11T10:00:00Z": time.Date(2024, 3, 11, 10, 0, 0, 0, time.UTC),
			"2024-03-11T10:00:00":  time.Date(2024, 3, 11, 10, 0, 0, 0, time.UTC),
		}
		for i, v := range cases {
			t.Run(i, func(t *testing.T) {
				var got time.Time
				in := fmt.Appendf(nil, "%q", i)
				err := json.Unmarshal(in, &got, json.WithUnmarshalers(unmarshalers))
				assert.NoError(t, err)
				assert.True(t, v.Equal(got))
			})
		}
	})

	t.Run("invalid", func(t *testing.T) {
		cases := []string{
			"123456789",    // non string token
			`"not-a-time"`, // invalid string
		}
		for _, s := range cases {
			t.Run(s, func(t *testing.T) {
				var got time.Time
				err := json.Unmarshal([]byte(s), &got, json.WithUnmarshalers(unmarshalers))
				assert.Error(t, err)
			})
		}
	})
}

func TestUnmarshalTemplateToken(t *testing.T) {
	type testcase struct {
		input    string
		expected TemplateToken
	}

	t.Run("scala", func(t *testing.T) {
		cases := map[string]testcase{
			"string": {
				input: `{"type": 0, "lit": "world"}`,
				expected: TemplateToken{
					Type:   TokenTypeString,
					String: "world",
				},
			},
			"number": {
				input: `{"type": 5, "num": 123.45}`,
				expected: TemplateToken{
					Type:   TokenTypeNumber,
					Number: 123.45,
				},
			},
			"boolean": {
				input: `{"type": 6, "bool": true}`,
				expected: TemplateToken{
					Type:    TokenTypeBoolean,
					Boolean: true,
				},
			},
		}
		for name, tc := range cases {
			t.Run(name, func(t *testing.T) {
				var got TemplateToken
				err := json.Unmarshal([]byte(tc.input), &got)
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, got)
			})
		}
	})

	t.Run("sequence", func(t *testing.T) {
		input := `{"type": 1, "seq": [{"type": 0, "lit": "repository"}]}`
		var got TemplateToken
		err := json.Unmarshal([]byte(input), &got)
		assert.NoError(t, err)
		assert.Equal(t, TokenTypeSequence, got.Type)
		assert.Len(t, got.Seq, 1)
		assert.Equal(t, TokenTypeString, got.Seq[0].Type)
		assert.Equal(t, "repository", got.Seq[0].String)
	})

	t.Run("mapping", func(t *testing.T) {
		input := `{"type": 2, "map": [{"key": {"type": 0, "lit": "repository"}, "value": {"type": 0, "lit": "dungdm93/shipyard"}}]}`
		var got TemplateToken
		err := json.Unmarshal([]byte(input), &got)
		assert.NoError(t, err)
		assert.Equal(t, TokenTypeMapping, got.Type)
		assert.Len(t, got.Map, 1)
		assert.Equal(t, "repository", got.Map[0].Key.String)
		assert.Equal(t, "dungdm93/shipyard", got.Map[0].Value.String)
	})

	t.Run("invalid", func(t *testing.T) {
		var token TemplateToken
		err := json.Unmarshal([]byte(`123`), &token)
		var target *json.SemanticError
		assert.ErrorAs(t, err, &target)
	})
}

func TestUnmarshalValue(t *testing.T) {
	unmarshalers := json.UnmarshalFromFunc(unmarshalValue)

	t.Run("simple", func(t *testing.T) {
		cases := map[string]Value{
			`"hello"`:            "hello",
			`123.45`:             123.45,
			`true`:               true,
			`null`:               nil,
			`["a", 3.14, false]`: []Value{"a", 3.14, false},
		}
		for in, expected := range cases {
			t.Run(in, func(t *testing.T) {
				var got Value
				err := json.Unmarshal([]byte(in), &got, json.WithUnmarshalers(unmarshalers))
				assert.NoError(t, err)
				assert.Equal(t, expected, got)
			})
		}
	})

	t.Run("typed", func(t *testing.T) {
		type testcase struct {
			input    string
			expected Value
		}
		cases := map[string]testcase{
			"string": {
				input:    `{"t": 0, "s": "world"}`,
				expected: Value("world"),
			},
			"bool": {
				input:    `{"t": 3, "b": true}`,
				expected: Value(true),
			},
			"number": {
				input:    `{"t": 4, "n": 123.45}`,
				expected: Value(123.45),
			},
			"array": {
				input:    `{"t": 1, "a": ["item1", true, 1.23]}`,
				expected: []Value{"item1", true, 1.23},
			},
			"array2": {
				input:    `{"t": 1, "a": [{"t": 0, "s": "item1"}, {"t": 3, "b": true}, {"t": 4, "n": 1.23}]}`,
				expected: []Value{"item1", true, 1.23},
			},
			"array_nested": {
				input: `{"t": 1, "a": [{"t": 1, "a": [{"t": 0, "s": "nested"}]}]}`,
				expected: []Value{
					[]Value{"nested"},
				},
			},
			"map": {
				input: `{"t": 2, "d": [{"k": "key", "v": "val"}]}`,
				expected: map[string]Value{
					"key": "val",
				},
			},
			"map2": {
				input: `{"t": 2, "d": [{"k": "a", "v": {"t": 0, "s": "1"}}, {"k": "b", "v": {"t": 3, "b": true}}]}`,
				expected: map[string]Value{
					"a": "1",
					"b": true,
				},
			},
		}
		for in, tc := range cases {
			t.Run(in, func(t *testing.T) {
				var got Value
				err := json.Unmarshal([]byte(tc.input), &got, json.WithUnmarshalers(unmarshalers))
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, got)
			})
		}
	})

	t.Run("invalid", func(t *testing.T) {
		cases := []string{
			`{"t": 99}`, // unknown type
		}
		for _, in := range cases {
			t.Run(in, func(t *testing.T) {
				var got Value
				err := json.Unmarshal([]byte(in), &got, json.WithUnmarshalers(unmarshalers))
				assert.Error(t, err)
			})
		}
	})
}
