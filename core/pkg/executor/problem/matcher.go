package problem

type Problem struct {
	File     string
	Line     string
	Column   string
	Severity string
	Code     string
	Message  string
	FromPath string
}

type Matcher struct {
}

func (m *Matcher) Match(line string) *Problem {
	return nil
}
