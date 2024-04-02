package contexts

type Context struct {
	Github    Github                         `json:"github" yaml:"github"`
	Env       map[string]string              `json:"env" yaml:"env"`
	Variables map[string]string              `json:"vars" yaml:"vars"`
	Job       Job                            `json:"job" yaml:"job"`
	Jobs      map[string]JobReusableWorkflow `json:"jobs" yaml:"jobs"`
	Steps     map[string]Step                `json:"steps" yaml:"steps"`
	Runner    Runner                         `json:"runner" yaml:"runner"`
	Secrets   map[string]string              `json:"secrets" yaml:"secrets"`
	Strategy  Strategy                       `json:"strategy" yaml:"strategy"`
	Matrix    map[string]string              `json:"matrix" yaml:"matrix"`
	Needs     map[string]Need                `json:"needs" yaml:"needs"`
	Inputs    map[string]any                 `json:"inputs" yaml:"inputs"`
}

type Strategy struct {
	FailFast    bool  `json:"fail-fast" yaml:"fail-fast"`
	JobIndex    int64 `json:"job-index" yaml:"job-index"`
	JobTotal    int64 `json:"job-total" yaml:"job-total"`
	MaxParallel int64 `json:"max-parallel" yaml:"max-parallel"`
}

type Need struct {
	Outputs map[string]string `json:"outputs" yaml:"outputs"`
	Result  JobResult         `json:"result" yaml:"result"`
}
