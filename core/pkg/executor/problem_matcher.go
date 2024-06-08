package executor

type ProblemMatchers struct {
	ProblemMatcher []ProblemMatcher `json:"problemMatcher,omitempty"`
}

type ProblemMatcher struct {
	Owner    string    `json:"owner,omitempty"`
	Severity string    `json:"severity,omitempty"`
	Pattern  []Pattern `json:"pattern,omitempty"`
}

type Pattern struct {
	File     *int   `json:"file,omitempty"`
	Line     *int   `json:"line,omitempty"`
	Column   *int   `json:"column,omitempty"`
	Severity *int   `json:"severity,omitempty"`
	Code     *int   `json:"code,omitempty"`
	Message  *int   `json:"message,omitempty"`
	FromPath *int   `json:"fromPath,omitempty"`
	Loop     bool   `json:"loop,omitempty"`
	Regexp   string `json:"regexp,omitempty"`
}
