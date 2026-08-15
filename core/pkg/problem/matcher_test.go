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

func TestNewMatcher(t *testing.T) {
	t.Run("empty-patterns", func(tt *testing.T) {
		m, err := NewMatcher("error", nil)
		assert.EqualError(tt, err, "patterns must not be empty")
		assert.Nil(tt, m)

		m, err = NewMatcher("error", []Pattern{})
		assert.EqualError(tt, err, "patterns must not be empty")
		assert.Nil(tt, m)
	})

	t.Run("single-line", func(tt *testing.T) {
		m, err := NewMatcher("warning", []Pattern{
			{Regexp: `^(.+)$`, Message: new(1)},
		})
		assert.NoError(tt, err)
		assert.NotNil(tt, m)
		assert.Equal(tt, 1, m.Len())

		m, err = NewMatcher("warning", []Pattern{
			{Regexp: `[invalid`, Message: new(1)},
		})
		assert.Error(tt, err)
		assert.Nil(tt, m)
	})

	t.Run("multi-line", func(tt *testing.T) {
		m, err := NewMatcher("warning", []Pattern{
			{Regexp: `^file:(.+)$`, File: new(1)},
			{Regexp: `^msg:(.+)$`, Message: new(1)},
		})
		assert.NoError(tt, err)
		assert.NotNil(tt, m)
		assert.Equal(tt, 2, m.Len())

		m, err = NewMatcher("warning", []Pattern{
			{Regexp: `^file:(.+)$`, File: new(1)},
			{Regexp: `[invalid`, Message: new(1)},
		})
		assert.Error(tt, err)
		assert.Nil(tt, m)
	})
}

func TestSingleLineMatcher(t *testing.T) {
	t.Run("default-severity", func(tt *testing.T) {
		matcher, err := newSingleLineMatcher("notice", Pattern{
			Regexp: `^(.+)$`, Message: new(1),
		})
		assert.NoError(tt, err)
		assert.Equal(tt, 1, matcher.Len())

		assert.EqualValues(tt, matcher.Match(nil, "just-a-notice"), &Problem{Severity: "notice", Message: "just-a-notice"})
	})

	t.Run("no-match", func(tt *testing.T) {
		matcher, err := newSingleLineMatcher("notice", Pattern{
			Regexp: `^notice: (.+)$`, Message: new(1),
		})
		assert.NoError(tt, err)
		assert.Nil(tt, matcher.Match(nil, "unrelated text"))
	})
}

func TestMultiLineMatcher(t *testing.T) {
	t.Run("default-severity", func(tt *testing.T) {
		patterns := []Pattern{
			{Regexp: `^(ERROR)?(?: )?(.+):$`, Severity: new(1), Code: new(2)},
			{Regexp: `^(.+)$`, Message: new(1)},
		}
		matcher, err := newMultiLineMatcher("warning", patterns)
		assert.NoError(tt, err)
		assert.Equal(tt, 2, matcher.Len())
		states := make([]*Problem, matcher.Len()-1)

		assert.Nil(tt, matcher.Match(states, "ABC:"))
		assert.EqualValues(tt, matcher.Match(states, "not-working"), &Problem{Severity: "warning", Code: "ABC", Message: "not-working"})

		assert.Nil(tt, matcher.Match(states, "ERROR ABC:"))
		assert.EqualValues(tt, matcher.Match(states, "not-working"), &Problem{Severity: "ERROR", Code: "ABC", Message: "not-working"})
	})

	t.Run("loop", func(tt *testing.T) {
		testMultiLineMatcher(tt, true)
	})
	t.Run("nonloop", func(tt *testing.T) {
		testMultiLineMatcher(tt, false)
	})
}

func testMultiLineMatcher(t *testing.T, loop bool) {
	t.Run("accumulate", func(tt *testing.T) {
		patterns := []Pattern{
			{Regexp: `^(.+)$`, File: new(1)},
			{Regexp: `^(.+)$`, Code: new(1)},
			{Regexp: `^message:(.+)$`, Message: new(1), Loop: loop},
		}
		matcher, err := newMultiLineMatcher("", patterns)
		assert.NoError(tt, err)
		assert.Equal(tt, 3, matcher.Len())
		states := make([]*Problem, matcher.Len()-1)

		assert.Nil(tt, matcher.Match(states, "file1"))
		assert.Nil(tt, matcher.Match(states, "code1"))
		assert.EqualValues(tt, matcher.Match(states, "message:message1"), &Problem{File: "file1", Code: "code1", Message: "message1"})
		if loop {
			assert.EqualValues(tt, matcher.Match(states, "message:message1-2"), &Problem{File: "file1", Code: "code1", Message: "message1-2"}) // loop
		}

		assert.Nil(tt, matcher.Match(states, "abc")) // discarded
		assert.Nil(tt, matcher.Match(states, "file2"))
		assert.Nil(tt, matcher.Match(states, "code2"))
		assert.EqualValues(tt, matcher.Match(states, "message:message2"), &Problem{File: "file2", Code: "code2", Message: "message2"})

		assert.Nil(tt, matcher.Match(states, "abc")) // discarded
		assert.Nil(tt, matcher.Match(states, "abc")) // discarded
		assert.Nil(tt, matcher.Match(states, "file3"))
		assert.Nil(tt, matcher.Match(states, "code3"))
		assert.EqualValues(tt, matcher.Match(states, "message:message3"), &Problem{File: "file3", Code: "code3", Message: "message3"})
		if !loop {
			assert.Nil(tt, matcher.Match(states, "message:message3"))
		}
	})

	t.Run("broken_match", func(tt *testing.T) {
		patterns := []Pattern{
			{Regexp: `^(.+)$`, File: new(1)},
			{Regexp: `^(.+)$`, Severity: new(1)},
			{Regexp: `^message:(.+)$`, Message: new(1), Loop: loop},
		}
		matcher, err := newMultiLineMatcher("", patterns)
		assert.NoError(tt, err)
		assert.Equal(tt, 3, matcher.Len())
		states := make([]*Problem, matcher.Len()-1)

		assert.Nil(tt, matcher.Match(states, "my-file.cs"))
		assert.Nil(tt, matcher.Match(states, "real-bad"))
		assert.EqualValues(tt, matcher.Match(states, "message:not-working"), &Problem{File: "my-file.cs", Severity: "real-bad", Message: "not-working"})
		if loop {
			assert.EqualValues(tt, matcher.Match(states, "message:problem"), &Problem{File: "my-file.cs", Severity: "real-bad", Message: "problem"})
		}

		assert.Nil(tt, matcher.Match(states, "other-file.cs"))    // file - breaks the loop
		assert.Nil(tt, matcher.Match(states, "message:not-good")) // severity - also matches the message pattern, therefore guarantees sufficient previous state has been cleared
		assert.EqualValues(tt, matcher.Match(states, "message:broken"), &Problem{File: "other-file.cs", Severity: "message:not-good", Message: "broken"})
	})

	t.Run("extracts_props", func(tt *testing.T) {
		patterns := []Pattern{
			{Regexp: `^file:(.+) fromPath:(.+)$`, File: new(1), FromPath: new(2)},
			{Regexp: `^severity:(.+)$`, Severity: new(1)},
			{Regexp: `^line:(.+) column:(.+) code:(.+) message:(.+)$`,
				Line: new(1), Column: new(2), Code: new(3), Message: new(4), Loop: loop},
		}
		matcher, err := newMultiLineMatcher("", patterns)
		assert.NoError(tt, err)
		assert.Equal(tt, 3, matcher.Len())
		states := make([]*Problem, matcher.Len()-1)

		assert.Nil(tt, matcher.Match(states, "file:my-file.cs fromPath:my-project.proj"))
		assert.Nil(tt, matcher.Match(states, "severity:real-bad"))
		assert.EqualValues(tt, matcher.Match(states, "line:123 column:45 code:uh-oh message:not-working"), &Problem{
			File:     "my-file.cs",
			FromPath: "my-project.proj",
			Severity: "real-bad",
			Line:     "123",
			Column:   "45",
			Code:     "uh-oh",
			Message:  "not-working",
		})
		if !loop {
			assert.Nil(tt, matcher.Match(states, "line:234 column:56 code:yikes message:broken"))
			return
		}
		assert.EqualValues(tt, matcher.Match(states, "line:234 column:56 code:yikes message:broken"), &Problem{
			File:     "my-file.cs",
			FromPath: "my-project.proj",
			Severity: "real-bad",
			Line:     "234",
			Column:   "56",
			Code:     "yikes",
			Message:  "broken",
		})
		assert.EqualValues(tt, matcher.Match(states, "line:345 column:67 code:failed message:cant-do-that"), &Problem{
			File:     "my-file.cs",
			FromPath: "my-project.proj",
			Severity: "real-bad",
			Line:     "345",
			Column:   "67",
			Code:     "failed",
			Message:  "cant-do-that",
		})
	})
}
