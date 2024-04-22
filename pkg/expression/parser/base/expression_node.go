package base

import (
	"github.com/dungdm93/drasi/pkg/expression/interfaces"
)

type (
	ExpressionNodeBase struct {
		Container interfaces.IContainer
		interfaces.IExpressionNode
		Level int
		// Name is used for tracing. Normally the parser will set the Name. However, if a node
		// is added manually, then the Name may not be set and will fall back to the type Name.
		Name string
	}
)
