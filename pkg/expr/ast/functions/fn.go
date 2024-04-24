package functions

import (
	"github.com/dungdm93/drasi/pkg/expr/ast/base"
	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

// Fn represent keyword available from parseContext eg: github, job, steps, env
type Fn struct {
	interfaces.Fn
	params []interfaces.Node
	base.Node
}

// AddParameter add node as current container's params
func (f *Fn) AddParameter(node interfaces.Node) {
	f.params = append(f.params, node)
	node.SetCtn(f)
}

// Parameters return values of all parameter of this Container node.
// This is read-only, so we will return a slice of value
func (f *Fn) Parameters() []interfaces.Node {
	var result []interfaces.Node
	for _, p := range f.params {
		result = append(result, p)
	}
	return result
}

func (f *Fn) GetLevel() (level int) {
	return f.Level
}
