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
			problem.File = result[i]
		})
	}
	if p.Line != nil {
		i := *p.Line
		sr = append(sr, func(result []string, problem *Problem) {
			problem.Line = result[i]
		})
	}
	if p.Column != nil {
		i := *p.Column
		sr = append(sr, func(result []string, problem *Problem) {
			problem.Column = result[i]
		})
	}
	if p.Severity != nil {
		i := *p.Severity
		sr = append(sr, func(result []string, problem *Problem) {
			problem.Severity = result[i]
		})
	}
	if p.Code != nil {
		i := *p.Code
		sr = append(sr, func(result []string, problem *Problem) {
			problem.Code = result[i]
		})
	}
	if p.Message != nil {
		i := *p.Message
		sr = append(sr, func(result []string, problem *Problem) {
			problem.Message = result[i]
		})
	}
	if p.FromPath != nil {
		i := *p.FromPath
		sr = append(sr, func(result []string, problem *Problem) {
			problem.FromPath = result[i]
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
