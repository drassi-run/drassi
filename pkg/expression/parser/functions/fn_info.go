package functions

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

func (n *FnInfo[T]) GetName() string {
	return n.name
}

func (n *FnInfo[T]) MinParameters() int {
	return n.minParameters
}

func (n *FnInfo[T]) MaxParameters() int {
	return n.maxParameters
}

// CreateNode create a node with of type T that is embed type Fn
func (n *FnInfo[T]) CreateNode() any {
	return new(T)
}
