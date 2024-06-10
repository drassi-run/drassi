package functions

import (
	"github.com/dungdm93/drassi/core/pkg/expression/ast/ast_ifaces"
	"github.com/dungdm93/drassi/core/pkg/expression/ast/base"
)

// Fn represent keyword available from parseContext eg: github, job, steps, env
type Fn struct {
	ast_ifaces.Fn
	params []ast_ifaces.ExprNode
	base.Node
}

// AddParameter add node as current container's params
func (f *Fn) AddParameter(node ast_ifaces.ExprNode) {
	f.params = append(f.params, node)
	node.SetCtn(f)
}

// Parameters return values of all parameter of this Container node.
// This is read-only, so we will return a slice of value
func (f *Fn) Parameters() []ast_ifaces.ExprNode {
	var result []ast_ifaces.ExprNode
	for _, p := range f.params {
		result = append(result, p)
	}
	return result
}

func (f *Fn) GetLevel() (level int) {
	return f.Level
}
