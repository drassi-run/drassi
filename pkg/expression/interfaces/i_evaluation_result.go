package interfaces

type IEvaluationResult interface {
	Raw() any
	Value() any
	GetKind() ValueKind
	Level() int
	IsFalsy() bool
	IsTruthy() bool
	AbstractEqual(right IEvaluationResult) bool
	AbstractGreaterThan(right IEvaluationResult) bool
	AbstractGreaterThanOrEqual(right IEvaluationResult) bool
	AbstractLessThan(right IEvaluationResult) bool
	AbstractLessThanOrEqual(right IEvaluationResult) bool
	AbstractNotEqual(right IEvaluationResult) bool
	ConvertToNumber() float64
	ConvertToString() string
	IsPrimitive() bool
	TryGetCollectionInterface() (ok bool, collection any)
}
