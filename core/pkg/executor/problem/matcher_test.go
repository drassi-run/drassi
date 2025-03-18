/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package problem

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSingleLineMatcher(t *testing.T) {
	t.Run("default-severity", func(tt *testing.T) {
		matcher, err := newSingleLineMatcher("notice", Pattern{
			Regexp: `^(.+)$`, Message: pointer(1),
		})
		assert.NoError(tt, err)

		assert.EqualValues(tt, matcher.Match("just-a-notice"), &Problem{Severity: "notice", Message: "just-a-notice"})
	})
}

func TestMultiLineMatcher(t *testing.T) {
	t.Run("default-severity", func(tt *testing.T) {
		patterns := []Pattern{
			{Regexp: `^(ERROR)?(?: )?(.+):$`, Severity: pointer(1), Code: pointer(2)},
			{Regexp: `^(.+)$`, Message: pointer(1)},
		}
		matcher, err := newMultiLineMatcher("warning", patterns)
		assert.NoError(tt, err)

		assert.Nil(tt, matcher.Match("ABC:"))
		assert.EqualValues(tt, matcher.Match("not-working"), &Problem{Severity: "warning", Code: "ABC", Message: "not-working"})

		assert.Nil(tt, matcher.Match("ERROR ABC:"))
		assert.EqualValues(tt, matcher.Match("not-working"), &Problem{Severity: "ERROR", Code: "ABC", Message: "not-working"})
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
			{Regexp: `^(.+)$`, File: pointer(1)},
			{Regexp: `^(.+)$`, Code: pointer(1)},
			{Regexp: `^message:(.+)$`, Message: pointer(1), Loop: loop},
		}
		matcher, err := newMultiLineMatcher("", patterns)
		assert.NoError(tt, err)

		assert.Nil(tt, matcher.Match("file1"))
		assert.Nil(tt, matcher.Match("code1"))
		assert.EqualValues(tt, matcher.Match("message:message1"), &Problem{File: "file1", Code: "code1", Message: "message1"})
		if loop {
			assert.EqualValues(tt, matcher.Match("message:message1-2"), &Problem{File: "file1", Code: "code1", Message: "message1-2"}) // loop
		}

		assert.Nil(tt, matcher.Match("abc")) // discarded
		assert.Nil(tt, matcher.Match("file2"))
		assert.Nil(tt, matcher.Match("code2"))
		assert.EqualValues(tt, matcher.Match("message:message2"), &Problem{File: "file2", Code: "code2", Message: "message2"})

		assert.Nil(tt, matcher.Match("abc")) // discarded
		assert.Nil(tt, matcher.Match("abc")) // discarded
		assert.Nil(tt, matcher.Match("file3"))
		assert.Nil(tt, matcher.Match("code3"))
		assert.EqualValues(tt, matcher.Match("message:message3"), &Problem{File: "file3", Code: "code3", Message: "message3"})
		if !loop {
			assert.Nil(tt, matcher.Match("message:message3"))
		}
	})

	t.Run("broken_match", func(tt *testing.T) {
		patterns := []Pattern{
			{Regexp: `^(.+)$`, File: pointer(1)},
			{Regexp: `^(.+)$`, Severity: pointer(1)},
			{Regexp: `^message:(.+)$`, Message: pointer(1), Loop: loop},
		}
		matcher, err := newMultiLineMatcher("", patterns)
		assert.NoError(tt, err)

		assert.Nil(tt, matcher.Match("my-file.cs"))
		assert.Nil(tt, matcher.Match("real-bad"))
		assert.EqualValues(tt, matcher.Match("message:not-working"), &Problem{File: "my-file.cs", Severity: "real-bad", Message: "not-working"})
		if loop {
			assert.EqualValues(tt, matcher.Match("message:problem"), &Problem{File: "my-file.cs", Severity: "real-bad", Message: "problem"})
		}

		assert.Nil(tt, matcher.Match("other-file.cs"))    // file - breaks the loop
		assert.Nil(tt, matcher.Match("message:not-good")) // severity - also matches the message pattern, therefore guarantees sufficient previous state has been cleared
		assert.EqualValues(tt, matcher.Match("message:broken"), &Problem{File: "other-file.cs", Severity: "message:not-good", Message: "broken"})
	})

	t.Run("extracts_props", func(tt *testing.T) {
		patterns := []Pattern{
			{Regexp: `^file:(.+) fromPath:(.+)$`, File: pointer(1), FromPath: pointer(2)},
			{Regexp: `^severity:(.+)$`, Severity: pointer(1)},
			{Regexp: `^line:(.+) column:(.+) code:(.+) message:(.+)$`,
				Line: pointer(1), Column: pointer(2), Code: pointer(3), Message: pointer(4), Loop: loop},
		}
		matcher, err := newMultiLineMatcher("", patterns)
		assert.NoError(tt, err)

		assert.Nil(tt, matcher.Match("file:my-file.cs fromPath:my-project.proj"))
		assert.Nil(tt, matcher.Match("severity:real-bad"))
		assert.EqualValues(tt, matcher.Match("line:123 column:45 code:uh-oh message:not-working"), &Problem{
			File:     "my-file.cs",
			FromPath: "my-project.proj",
			Severity: "real-bad",
			Line:     "123",
			Column:   "45",
			Code:     "uh-oh",
			Message:  "not-working",
		})
		if !loop {
			assert.Nil(tt, matcher.Match("line:234 column:56 code:yikes message:broken"))
			return
		}
		assert.EqualValues(tt, matcher.Match("line:234 column:56 code:yikes message:broken"), &Problem{
			File:     "my-file.cs",
			FromPath: "my-project.proj",
			Severity: "real-bad",
			Line:     "234",
			Column:   "56",
			Code:     "yikes",
			Message:  "broken",
		})
		assert.EqualValues(tt, matcher.Match("line:345 column:67 code:failed message:cant-do-that"), &Problem{
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
