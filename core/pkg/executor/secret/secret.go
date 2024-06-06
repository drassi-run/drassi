package secret

import (
	"regexp"
	"strings"
)

// Position represent the location of a secret at [Start, End) in a string
type Position struct {
	Start int
	End   int
}

type Secret interface {
	At(input string) []*Position
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTLogging/Logging/ValueSecret.cs
type valueSecret struct {
	val string
}

func NewValueSecret(val string) Secret {
	return &valueSecret{val}
}

func (s *valueSecret) At(input string) []*Position {
	pos := make([]*Position, 0)

	for idx := s.indexAt(input, 0); idx > -1; idx = s.indexAt(input, idx+1) {
		p := &Position{
			Start: idx,
			End:   idx + len(s.val),
		}
		pos = append(pos, p)
	}
	return pos
}

func (s *valueSecret) indexAt(input string, n int) int {
	idx := strings.Index(input[n:], s.val)
	if idx > -1 {
		idx += n
	}
	return idx
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTLogging/Logging/RegexSecret.cs
type regexSecret struct {
	re *regexp.Regexp
}

func NewRegexSecret(re *regexp.Regexp) Secret {
	return &regexSecret{re}
}

func (s *regexSecret) At(input string) []*Position {
	pos := make([]*Position, 0)

	for p := s.findAt(input, 0); p != nil; p = s.findAt(input, p.Start+1) {
		pos = append(pos, p)
	}
	return pos
}

// we use loop with re.FindStringIndex because
// regexp.FindAll* method return non-overlapping matches
func (s *regexSecret) findAt(input string, n int) *Position {
	loc := s.re.FindStringIndex(input[n:])
	// not match or match empty string
	if len(loc) < 2 || loc[0] >= loc[1] {
		return nil
	}
	return &Position{
		Start: loc[0] + n,
		End:   loc[1] + n,
	}
}
