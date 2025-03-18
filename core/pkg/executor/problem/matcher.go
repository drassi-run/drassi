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
	Match(line string) *Problem
	Reset()
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

func (m *singleLineMatcher) Match(line string) *Problem {
	if r := m.extractor.Match(line); r == nil {
		return nil
	} else {
		p := &Problem{Severity: m.severity}
		m.extractor.Extract(r, p)
		return p
	}
}

func (m *singleLineMatcher) Reset() {}

type multiLineMatcher struct {
	severity   string
	extractors []*extractor
	loop       bool

	states []*Problem
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
	states := make([]*Problem, len(patterns)-1)

	matcher := &multiLineMatcher{
		severity:   severity,
		extractors: extractors,
		loop:       loop,
		states:     states,
	}
	return matcher, nil
}

func (m *multiLineMatcher) Match(line string) *Problem {
	// NOTE: len(states) = len(extractors)-1 = last extractors idx
	for i := len(m.states); i >= 0; i-- {
		var runningMatch *Problem = nil
		if i > 0 {
			// Previous pattern's not matched
			if runningMatch = m.states[i-1]; runningMatch == nil {
				continue
			}
			m.states[i-1] = nil
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

		// Not the last patten
		if i != len(m.states) {
			m.states[i] = runningMatch
			continue
		}

		// The last pattern
		m.Reset()
		if m.loop {
			p := *runningMatch // clone Problem
			m.states[i-1] = &p
		}
		return runningMatch
	}
	return nil
}

func (m *multiLineMatcher) Reset() {
	clear(m.states)
}
