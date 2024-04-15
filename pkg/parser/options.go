package parser

type EvaluationOption struct {
	maxMemory int
}

func (e *EvaluationOption) GetMaxMemory() int {
	return e.maxMemory
}

func (e *EvaluationOption) SetMaxMemory(m int) {
	e.maxMemory = m
}
