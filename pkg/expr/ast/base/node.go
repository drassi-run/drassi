package base

import (
	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

type (
	Node struct {
		Ctn interfaces.Container
		interfaces.Node
		Level int
		// Name is used for tracing. Normally the ast will set the Name. However, if a node
		// is added manually, then the Name may not be set and will fall back to the type Name.
		Name string
	}
)
