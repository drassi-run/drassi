package parser

type INamedValueInfo[T any] interface {
	GetName() string
	CreateNode() any
}

type NamedValueInfo[T any] struct {
	name string
}

func NewNamedValueInfo[T any](name string) *NamedValueInfo[T] {
	return &NamedValueInfo[T]{
		name: name,
	}
}

func (n *NamedValueInfo[T]) GetName() string {
	return n.name
}

func (n *NamedValueInfo[T]) CreateNode() any {
	return new(T)
}
