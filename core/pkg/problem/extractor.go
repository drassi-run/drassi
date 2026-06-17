/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package problem

import "regexp"

type extractor struct {
	re *regexp.Regexp
	sr []func([]string, *Problem)
}

func set(v string, val *string) {
	if v != "" {
		*val = v
	}
}

func newExtractor(p Pattern) (*extractor, error) {
	re, err := regexp.Compile(p.Regexp)
	if err != nil {
		return nil, err
	}
	proc := &extractor{re: re}

	if p.File != nil {
		i := *p.File
		proc.sr = append(proc.sr, func(result []string, problem *Problem) {
			set(result[i], &problem.File)
		})
	}
	if p.Line != nil {
		i := *p.Line
		proc.sr = append(proc.sr, func(result []string, problem *Problem) {
			set(result[i], &problem.Line)
		})
	}
	if p.Column != nil {
		i := *p.Column
		proc.sr = append(proc.sr, func(result []string, problem *Problem) {
			set(result[i], &problem.Column)
		})
	}
	if p.Severity != nil {
		i := *p.Severity
		proc.sr = append(proc.sr, func(result []string, problem *Problem) {
			set(result[i], &problem.Severity)
		})
	}
	if p.Code != nil {
		i := *p.Code
		proc.sr = append(proc.sr, func(result []string, problem *Problem) {
			set(result[i], &problem.Code)
		})
	}
	if p.Message != nil {
		i := *p.Message
		proc.sr = append(proc.sr, func(result []string, problem *Problem) {
			set(result[i], &problem.Message)
		})
	}
	if p.FromPath != nil {
		i := *p.FromPath
		proc.sr = append(proc.sr, func(result []string, problem *Problem) {
			set(result[i], &problem.FromPath)
		})
	}

	return proc, nil
}

func (p *extractor) Match(line string) []string {
	return p.re.FindStringSubmatch(line)
}

func (p *extractor) Extract(result []string, obj *Problem) {
	for _, f := range p.sr {
		f(result, obj)
	}
}
