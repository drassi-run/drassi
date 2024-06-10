package base

import (
	"github.com/dungdm93/drassi/core/pkg/expression/ast/ast_ifaces"
)

type (
	Node struct {
		Ctn ast_ifaces.Container
		ast_ifaces.ExprNode
		Level int
		// Name is used for tracing. Normally the ast will set the Name. However, if a node
		// is added manually, then the Name may not be set and will fall back to the type Name.
		Name string
	}
)
