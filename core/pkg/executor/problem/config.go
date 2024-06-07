package problem

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/IssueMatcher.cs#L230
type Configs struct {
	ProblemMatcher []Config `json:"problemMatcher,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/IssueMatcher.cs#L274
type Config struct {
	Owner    string    `json:"owner,omitempty"`
	Severity string    `json:"severity,omitempty"`
	Pattern  []Pattern `json:"pattern,omitempty"`
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/IssueMatcher.cs#L398
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
