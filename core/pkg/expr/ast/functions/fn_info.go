package functions

import (
	"math"
)

type IFnInfo[T any] interface {
	GetName() string
	CreateNode() any
	MinParameters() int
	MaxParameters() int
}

type FnInfo[T any] struct {
	name          string
	minParameters int
	maxParameters int
}

func NewFunctionInfo[T any](name string, minParameters, maxParameters int) *FnInfo[T] {
	return &FnInfo[T]{
		name:          name,
		minParameters: minParameters,
		maxParameters: maxParameters,
	}
}

func (f *FnInfo[T]) GetName() string {
	return f.name
}

func (f *FnInfo[T]) MinParameters() int {
	return f.minParameters
}

func (f *FnInfo[T]) MaxParameters() int {
	return f.maxParameters
}

// CreateNode create a node with of type T that is embed type Fn
func (f *FnInfo[T]) CreateNode() any {
	return new(T)
}

func ContextBased(name string) any  {
	switch name {
	case "always":
		return NewFunctionInfo[Always]("always", 0, math.MaxInt32)
	case "cancelled":
		return NewFunctionInfo[Cancelled]("cancelled", 0, 0)
	case "success":
		return	NewFunctionInfo[Success]("success", 0, 0)
	case "failure":
		return NewFunctionInfo[Failure]("failure", 0, 0)
	case "hashfile":
		return NewFunctionInfo[HashFile]("hashfile", 1, math.MaxUint8)
	}
	return nil
}
