/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package problem

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	t.Run("owner/distinct", func(tt *testing.T) {
		conf := MatcherConfigs{
			Configs: []Config{
				{
					Owner: "myMatcher",
					Patterns: []Pattern{
						{Regexp: `^error: (.+)$`, Message: new(1)},
					},
				}, {
					Owner: "MYmatcher",
					Patterns: []Pattern{
						{Regexp: `^ERR: (.+)$`, Message: new(1)},
					},
				}},
		}

		err := conf.Validate()
		assert.Error(tt, err, "duplicate owner name MYmatcher")

		conf.Configs[0].Owner = "asdf"
		err = conf.Validate()
		assert.NoError(tt, err)
	})

	t.Run("owner/required", func(tt *testing.T) {
		conf := MatcherConfigs{
			Configs: []Config{{
				Patterns: []Pattern{
					{Regexp: `^error: (.+)$`, Message: new(1)},
				},
			}}}

		err := conf.Validate()
		assert.Error(tt, err, "owner must not be empty")

		conf.Configs[0].Owner = "asdf"
		err = conf.Validate()
		assert.NoError(tt, err)
	})

	t.Run("severity", func(tt *testing.T) {
		config := &Config{
			Owner: "myMatcher",
			Patterns: []Pattern{
				{Regexp: `^(.+)$`, Message: new(1)},
			},
		}
		err := config.Validate()
		assert.NoError(tt, err)

		config.Severity = "foobar"
		err = config.Validate()
		assert.Error(tt, err, "matcher myMatcher contains unexpected default severity: FOOBAR")

		config.Severity = "NOTICE"
		err = config.Validate()
		assert.NoError(tt, err)
	})

	t.Run("loop/single", func(tt *testing.T) {
		p1 := Pattern{
			Regexp:  "^error: (.+)$",
			Message: new(1),
			Loop:    true,
		}
		config := &Config{
			Owner:    "myMatcher",
			Patterns: []Pattern{p1},
		}
		err := config.Validate()
		assert.Error(tt, err, "only the last pattern in a multiline matcher may set 'loop'")

		p2 := Pattern{
			Regexp: "^file: (.+)$",
			File:   new(1),
		}
		config.Patterns = []Pattern{p2, p1}
		err = config.Validate()
		assert.NoError(tt, err)
	})

	t.Run("loop/last", func(tt *testing.T) {
		config := &Config{
			Owner: "myMatcher",
			Patterns: []Pattern{
				{Regexp: `^(error)$`, Severity: new(1)},
				{Regexp: `^file: (.+)$`, File: new(1), Loop: true},
				{Regexp: `^error: (.+)$`, Message: new(1), Loop: false},
			},
		}
		err := config.Validate()
		assert.Error(tt, err, "only the last pattern in a multiline matcher may set 'loop'")

		config.Patterns[1].Loop = false
		config.Patterns[2].Loop = true
		err = config.Validate()
		assert.NoError(tt, err)
	})

	t.Run("loop/message", func(tt *testing.T) {
		config := &Config{
			Owner: "myMatcher",
			Patterns: []Pattern{
				{Regexp: `^file: (.+)$`, Message: new(1)},
				{Regexp: `^file: (.+)$`, File: new(1), Loop: true},
			},
		}
		err := config.Validate()
		assert.Error(tt, err, "the loop pattern must set 'message'")

		config.Patterns[1].Loop = false
		err = config.Validate()
		assert.NoError(tt, err)
	})

	t.Run("message/first", func(tt *testing.T) {
		config := &Config{
			Owner: "myMatcher",
			Patterns: []Pattern{
				{Regexp: `^file: (.+)$`, Message: new(1)},
				{Regexp: `^error: (.+)$`, File: new(1)},
			},
		}
		err := config.Validate()
		assert.NoError(tt, err)
	})

	t.Run("message/required", func(tt *testing.T) {
		config := &Config{
			Owner: "myMatcher",
			Patterns: []Pattern{
				{Regexp: `^error: (.+)$`, File: new(1)},
			},
		}
		err := config.Validate()
		assert.Error(tt, err, "at least one pattern must set 'message'")

		config.Patterns[0].File = nil
		config.Patterns[0].Message = new(1)
		err = config.Validate()
		assert.NoError(tt, err)
	})

	t.Run("regexp/required", func(tt *testing.T) {
		config := &Config{
			Owner: "myMatcher",
			Patterns: []Pattern{
				{Message: new(1)},
			},
		}
		err := config.Validate()
		assert.Error(tt, err, "regexp must not be empty")

		config.Patterns[0].Regexp = `^error: (.+)$`
		err = config.Validate()
		assert.NoError(tt, err)
	})

	t.Run("property/duplicated", func(tt *testing.T) {
		config := &Config{
			Owner: "myMatcher",
			Patterns: []Pattern{
				{Regexp: `^severity: (.+)$`, File: new(1)},
				{Regexp: `^file: (.+)$`, File: new(1)},
				{Regexp: `^(.+)$`, Message: new(1)},
			},
		}
		err := config.Validate()
		assert.Error(tt, err, `the property "file" is set twice`)

		config.Patterns[0].File = nil
		config.Patterns[0].Severity = new(1)
		err = config.Validate()
		assert.NoError(tt, err)
	})

	t.Run("property/out-of-range", func(tt *testing.T) {
		config := &Config{
			Owner: "myMatcher",
			Patterns: []Pattern{
				{Regexp: `^(.+)$`, Message: new(2)},
			},
		}
		err := config.Validate()
		assert.Error(tt, err, `the value 2 of property "message" is out of range`)

		config.Patterns[0].Message = new(1)
		err = config.Validate()
		assert.NoError(tt, err)
	})

	t.Run("property/negative", func(tt *testing.T) {
		config := &Config{
			Owner: "myMatcher",
			Patterns: []Pattern{
				{Regexp: `^(.+)$`, Message: new(-10)},
			},
		}
		err := config.Validate()
		assert.Error(tt, err, "at least one pattern must set 'message'")

		config.Patterns[0].Message = new(1)
		err = config.Validate()
		assert.NoError(tt, err)
	})
}
