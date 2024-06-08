package problem

import "regexp"

type setter = func([]string, *Problem)

type extractor struct {
	re *regexp.Regexp
	sr []setter
}

func newExtractor(p Pattern) (*extractor, error) {
	re, err := regexp.Compile(p.Regexp)
	if err != nil {
		return nil, err
	}
	var sr []setter
	if p.File != nil {
		i := *p.File
		sr = append(sr, func(result []string, problem *Problem) {
			if v := result[i]; v != "" {
				problem.File = v
			}
		})
	}
	if p.Line != nil {
		i := *p.Line
		sr = append(sr, func(result []string, problem *Problem) {
			if v := result[i]; v != "" {
				problem.Line = v
			}
		})
	}
	if p.Column != nil {
		i := *p.Column
		sr = append(sr, func(result []string, problem *Problem) {
			if v := result[i]; v != "" {
				problem.Column = v
			}
		})
	}
	if p.Severity != nil {
		i := *p.Severity
		sr = append(sr, func(result []string, problem *Problem) {
			if v := result[i]; v != "" {
				problem.Severity = v
			}
		})
	}
	if p.Code != nil {
		i := *p.Code
		sr = append(sr, func(result []string, problem *Problem) {
			if v := result[i]; v != "" {
				problem.Code = v
			}
		})
	}
	if p.Message != nil {
		i := *p.Message
		sr = append(sr, func(result []string, problem *Problem) {
			if v := result[i]; v != "" {
				problem.Message = v
			}
		})
	}
	if p.FromPath != nil {
		i := *p.FromPath
		sr = append(sr, func(result []string, problem *Problem) {
			if v := result[i]; v != "" {
				problem.FromPath = v
			}
		})
	}

	proc := &extractor{
		re: re,
		sr: sr,
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
