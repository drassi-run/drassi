package interfaces

type Container interface {
	Node
	Parameters() []Node
	AddParameter(node Node)
}
