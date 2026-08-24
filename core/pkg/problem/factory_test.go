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

func TestDetectorFactory(t *testing.T) {
	t.Run("Add_valid_single_line", func(tt *testing.T) {
		factory := make(DetectorFactory)
		err := factory.Add(Config{
			Owner:    "single",
			Severity: "NOTICE",
			Patterns: []Pattern{
				{Regexp: `^notice: (.+)$`, Message: new(1)},
			},
		})
		assert.NoError(tt, err)
		assert.Contains(tt, factory, "single")
		assert.Equal(tt, 1, factory["single"].Len())
	})

	t.Run("Add_valid_multi_line", func(tt *testing.T) {
		factory := make(DetectorFactory)
		err := factory.Add(Config{
			Owner: "multi",
			Patterns: []Pattern{
				{Regexp: `^file: (.+)$`, File: new(1)},
				{Regexp: `^error: (.+)$`, Message: new(1)},
			},
		})
		assert.NoError(tt, err)
		assert.Contains(tt, factory, "multi")
		assert.Equal(tt, 2, factory["multi"].Len())
	})

	t.Run("Add_invalid", func(tt *testing.T) {
		factory := make(DetectorFactory)
		err := factory.Add(Config{
			Owner:    "invalid",
			Patterns: []Pattern{},
		})
		assert.Error(tt, err)
		assert.NotContains(tt, factory, "invalid")
	})

	t.Run("Remove", func(tt *testing.T) {
		factory := make(DetectorFactory)
		_ = factory.Add(Config{
			Owner:    "test",
			Patterns: []Pattern{{Regexp: `^(.+)$`, Message: new(1)}},
		})
		assert.Contains(tt, factory, "test")

		factory.Remove("test")
		assert.NotContains(tt, factory, "test")
	})
}

func TestDetector(t *testing.T) {
	t.Run("single_matcher", func(tt *testing.T) {
		factory := make(DetectorFactory)
		_ = factory.Add(Config{
			Owner:    "single",
			Severity: "NOTICE",
			Patterns: []Pattern{
				{Regexp: `^notice: (.+)$`, Message: new(1)},
			},
		})

		d := factory.NewDetector()
		assert.Nil(tt, d.Detect("random line"))
		assert.EqualValues(tt, d.Detect("notice: something happened"), &Problem{
			Severity: "NOTICE",
			Message:  "something happened",
		})
	})

	t.Run("multi_matcher_accumulate", func(tt *testing.T) {
		factory := make(DetectorFactory)
		_ = factory.Add(Config{
			Owner:    "multi",
			Severity: "ERROR",
			Patterns: []Pattern{
				{Regexp: `^file: (.+)$`, File: new(1)},
				{Regexp: `^line: (\d+)$`, Line: new(1)},
				{Regexp: `^error: (.+)$`, Message: new(1)},
			},
		})

		d := factory.NewDetector()
		assert.Nil(tt, d.Detect("file: main.go"))
		assert.Nil(tt, d.Detect("line: 42"))
		assert.EqualValues(tt, d.Detect("error: syntax error"), &Problem{
			Severity: "ERROR",
			File:     "main.go",
			Line:     "42",
			Message:  "syntax error",
		})
	})

	t.Run("multi_matcher_loop", func(tt *testing.T) {
		factory := make(DetectorFactory)
		_ = factory.Add(Config{
			Owner: "multi-loop",
			Patterns: []Pattern{
				{Regexp: `^file: (.+)$`, File: new(1)},
				{Regexp: `^msg: (.+)$`, Message: new(1), Loop: true},
			},
		})

		d := factory.NewDetector()
		assert.Nil(tt, d.Detect("file: test.js"))
		assert.EqualValues(tt, d.Detect("msg: first problem"), &Problem{
			File:    "test.js",
			Message: "first problem",
		})
		assert.EqualValues(tt, d.Detect("msg: second problem"), &Problem{
			File:    "test.js",
			Message: "second problem",
		})
		assert.Nil(tt, d.Detect("unrelated line")) // breaks loop
		assert.Nil(tt, d.Detect("msg: third problem"))
	})

	t.Run("multiple_matchers_reset_others_on_match", func(tt *testing.T) {
		factory := make(DetectorFactory)
		_ = factory.Add(Config{
			Owner: "matcherA",
			Patterns: []Pattern{
				{Regexp: `^file: (.+)$`, File: new(1)},
				{Regexp: `^code: (\d+)$`, Code: new(1)},
				{Regexp: `^msgA: (.+)$`, Message: new(1)},
			},
		})
		_ = factory.Add(Config{
			Owner:    "matcherB",
			Severity: "FATAL",
			Patterns: []Pattern{
				{Regexp: `^fatal: (.+)$`, Message: new(1)},
			},
		})

		d := factory.NewDetector()

		// Start partial match for matcherA
		assert.Nil(tt, d.Detect("file: app.go"))
		assert.Nil(tt, d.Detect("code: 101"))

		// Matcher B matches completely - this must reset matcherA's states
		assert.EqualValues(tt, d.Detect("fatal: out of memory"), &Problem{
			Severity: "FATAL",
			Message:  "out of memory",
		})

		// Now matcherA line 3 is received, but matcherA was reset, so it fails to complete
		assert.Nil(tt, d.Detect("msgA: some error"))

		// Starting matcherA from beginning works again
		assert.Nil(tt, d.Detect("file: app.go"))
		assert.Nil(tt, d.Detect("code: 101"))
		assert.EqualValues(tt, d.Detect("msgA: some error"), &Problem{
			File:    "app.go",
			Code:    "101",
			Message: "some error",
		})
	})

	t.Run("two_multiline_matchers_reset_each_other", func(tt *testing.T) {
		factory := make(DetectorFactory)
		_ = factory.Add(Config{
			Owner: "matcherA",
			Patterns: []Pattern{
				{Regexp: `^fileA: (.+)$`, File: new(1)},
				{Regexp: `^msgA: (.+)$`, Message: new(1)},
			},
		})
		_ = factory.Add(Config{
			Owner: "matcherB",
			Patterns: []Pattern{
				{Regexp: `^fileB: (.+)$`, File: new(1)},
				{Regexp: `^msgB: (.+)$`, Message: new(1)},
			},
		})

		d := factory.NewDetector()

		// Partial match A
		assert.Nil(tt, d.Detect("fileA: a.go"))

		// Matcher B starts and finishes
		assert.Nil(tt, d.Detect("fileB: b.go"))
		assert.EqualValues(tt, d.Detect("msgB: errB"), &Problem{
			File:    "b.go",
			Message: "errB",
		})

		// Matcher A was reset by B's match, so msgA does not complete
		assert.Nil(tt, d.Detect("msgA: errA"))
	})

	t.Run("add_matcher_in_the_middle", func(tt *testing.T) {
		factory := make(DetectorFactory)
		_ = factory.Add(Config{
			Owner:    "matcherA",
			Severity: "NOTICE",
			Patterns: []Pattern{
				{Regexp: `^notice: (.+)$`, Message: new(1)},
			},
		})

		d := factory.NewDetector()

		assert.Nil(tt, d.Detect("random line"))
		assert.EqualValues(tt, d.Detect("notice: before adding new matcher"), &Problem{
			Severity: "NOTICE",
			Message:  "before adding new matcher",
		})

		// Add multi-line matcher in the middle
		err := factory.Add(Config{
			Owner:    "matcherB",
			Severity: "ERROR",
			Patterns: []Pattern{
				{Regexp: `^file: (.+)$`, File: new(1)},
				{Regexp: `^line: (\d+)$`, Line: new(1)},
				{Regexp: `^error: (.+)$`, Message: new(1)},
			},
		})
		assert.NoError(tt, err)

		// Test multi-line detection with matcherB added after NewDetector()
		assert.Nil(tt, d.Detect("file: main.go"))
		assert.Nil(tt, d.Detect("line: 42"))
		assert.EqualValues(tt, d.Detect("error: syntax error"), &Problem{
			Severity: "ERROR",
			File:     "main.go",
			Line:     "42",
			Message:  "syntax error",
		})

		// Matcher A still works
		assert.EqualValues(tt, d.Detect("notice: after adding new matcher"), &Problem{
			Severity: "NOTICE",
			Message:  "after adding new matcher",
		})

		// Add another multi-line loop matcher in the middle
		err = factory.Add(Config{
			Owner: "matcherC",
			Patterns: []Pattern{
				{Regexp: `^fileC: (.+)$`, File: new(1)},
				{Regexp: `^msgC: (.+)$`, Message: new(1), Loop: true},
			},
		})
		assert.NoError(tt, err)

		assert.Nil(tt, d.Detect("fileC: loop.go"))
		assert.EqualValues(tt, d.Detect("msgC: first issue"), &Problem{
			File:    "loop.go",
			Message: "first issue",
		})
		assert.EqualValues(tt, d.Detect("msgC: second issue"), &Problem{
			File:    "loop.go",
			Message: "second issue",
		})

		// Partial match before adding new matcher, then complete after
		_ = factory.Add(Config{
			Owner: "matcherD",
			Patterns: []Pattern{
				{Regexp: `^fileD: (.+)$`, File: new(1)},
				{Regexp: `^lineD: (\d+)$`, Line: new(1)},
				{Regexp: `^errorD: (.+)$`, Message: new(1)},
			},
		})
		assert.Nil(tt, d.Detect("fileD: foo.go"))

		// Add new matcher while matcherD is partially matched
		_ = factory.Add(Config{
			Owner:    "matcherE",
			Severity: "WARNING",
			Patterns: []Pattern{
				{Regexp: `^warnE: (.+)$`, Message: new(1)},
			},
		})

		assert.Nil(tt, d.Detect("lineD: 100"))
		assert.EqualValues(tt, d.Detect("errorD: broken"), &Problem{
			File:    "foo.go",
			Line:    "100",
			Message: "broken",
		})

		// Verify matcherE works as well
		assert.EqualValues(tt, d.Detect("warnE: something"), &Problem{
			Severity: "WARNING",
			Message:  "something",
		})
	})
}
