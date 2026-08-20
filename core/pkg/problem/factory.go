/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package problem

type DetectorFactory map[string]Matcher

func NewDetectorFactory() DetectorFactory {
	return make(DetectorFactory)
}

func (f DetectorFactory) Add(config Config) error {
	if matcher, err := NewMatcher(config.Severity, config.Patterns); err != nil {
		return err
	} else {
		f[config.Owner] = matcher
		return nil
	}
}

func (f DetectorFactory) Remove(owner string) {
	delete(f, owner)
}

func (f DetectorFactory) NewDetector() Detector {
	states := make(map[string][]*Problem)
	for owner, matcher := range f {
		if l := matcher.Len(); l > 1 {
			states[owner] = make([]*Problem, l-1)
		}
	}
	return &detector{
		matchers: f,
		states:   states,
	}
}

type Detector interface {
	Detect(line string) (pbl *Problem)
}

type detector struct {
	matchers map[string]Matcher
	states   map[string][]*Problem
}

func (d *detector) Detect(line string) (pbl *Problem) {
	var owner string

	for o, m := range d.matchers {
		s := d.states[o]
		if p := m.Match(s, line); p != nil {
			owner, pbl = o, p
			break
		}
	}

	// Matched - then reset other matchers
	if pbl != nil {
		for o, s := range d.states {
			if o != owner {
				clear(s)
			}
		}
	}

	return
}
