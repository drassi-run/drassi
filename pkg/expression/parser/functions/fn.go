package functions

import (
	"github.com/dungdm93/drasi/pkg/expression"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type IFn interface {
	expression.IContainer
}

// Fn represent keyword available from parseContext eg: github, job, steps, env
type Fn struct {
	IFn
	params []expression.IExpNode
	base.ExpressionNodeBs
}

// AddParameter add node as current container's params
func (f *Fn) AddParameter(node expression.IExpNode) {
	f.params = append(f.params, node)
	node.SetContainer(f)
}

// Parameters return values of all parameter of this ContainerBs node.
// This is read-only, so we will return a slice of value
func (f *Fn) Parameters() []expression.IExpNode {
	var result []expression.IExpNode
	for _, p := range f.params {
		result = append(result, p)
	}
	return result
}

func (f *Fn) GetLevel() (level int) {
	return f.Level
}
