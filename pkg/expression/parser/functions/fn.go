package functions

import (
	"github.com/dungdm93/drasi/pkg/expression/interfaces"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type IFn interface {
	interfaces.IContainer
}

// Fn represent keyword available from parseContext eg: github, job, steps, env
type Fn struct {
	IFn
	params []interfaces.IExpressionNode
	base.ExpressionNodeBase
}

// AddParameter add node as current container's params
func (f *Fn) AddParameter(node interfaces.IExpressionNode) {
	f.params = append(f.params, node)
	node.SetContainer(f)
}

// Parameters return values of all parameter of this ContainerBase node.
// This is read-only, so we will return a slice of value
func (f *Fn) Parameters() []interfaces.IExpressionNode {
	var result []interfaces.IExpressionNode
	for _, p := range f.params {
		result = append(result, p)
	}
	return result
}

func (f *Fn) GetLevel() (level int) {
	return f.Level
}
