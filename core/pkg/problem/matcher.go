/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package problem

import "fmt"

type Problem struct {
	File     string
	Line     string
	Column   string
	Severity string
	Code     string
	Message  string
	FromPath string
}

type Matcher interface {
	Len() int
	Match(states []*Problem, line string) *Problem
}

func NewMatcher(severity string, patterns []Pattern) (Matcher, error) {
	switch len(patterns) {
	case 0:
		return nil, fmt.Errorf("patterns must not be empty")
	case 1:
		return newSingleLineMatcher(severity, patterns[0])
	default:
		return newMultiLineMatcher(severity, patterns)
	}
}

type singleLineMatcher struct {
	severity  string
	extractor *extractor
}

func newSingleLineMatcher(severity string, pattern Pattern) (Matcher, error) {
	x, err := newExtractor(pattern)
	if err != nil {
		return nil, err
	}

	matcher := &singleLineMatcher{
		severity:  severity,
		extractor: x,
	}
	return matcher, nil
}

func (m *singleLineMatcher) Len() int {
	return 1
}

func (m *singleLineMatcher) Match(_ []*Problem, line string) *Problem {
	if r := m.extractor.Match(line); r == nil {
		return nil
	} else {
		p := &Problem{Severity: m.severity}
		m.extractor.Extract(r, p)
		return p
	}
}

type multiLineMatcher struct {
	severity   string
	extractors []*extractor
	loop       bool
}

func newMultiLineMatcher(severity string, patterns []Pattern) (Matcher, error) {
	extractors := make([]*extractor, len(patterns))
	for i, p := range patterns {
		x, err := newExtractor(p)
		if err != nil {
			return nil, err
		}
		extractors[i] = x
	}
	loop := patterns[len(patterns)-1].Loop

	matcher := &multiLineMatcher{
		severity:   severity,
		extractors: extractors,
		loop:       loop,
	}
	return matcher, nil
}

func (m *multiLineMatcher) Len() int {
	return len(m.extractors)
}

func (m *multiLineMatcher) Match(states []*Problem, line string) *Problem {
	// NOTE: len(states) = len(extractors)-1 = last extractors idx
	for i := len(states); i >= 0; i-- {
		var runningMatch *Problem = nil
		if i > 0 {
			// Previous pattern's not matched
			if runningMatch = states[i-1]; runningMatch == nil {
				continue
			}
			states[i-1] = nil
		}

		xr := m.extractors[i]
		r := xr.Match(line)

		// Not matched
		if r == nil {
			continue
		}

		// Matched
		if runningMatch == nil {
			runningMatch = &Problem{Severity: m.severity}
		}
		xr.Extract(r, runningMatch)

		// Not the last pattern
		if i != len(states) {
			states[i] = runningMatch
			continue
		}

		// The last pattern
		clear(states)
		if m.loop {
			p := *runningMatch // clone Problem
			states[i-1] = &p
		}
		return runningMatch
	}
	return nil
}
