package parser

type iFn interface {
	iContainer
}

// Fn represent keyword available from parseContext eg: github, job, steps, env
type Fn struct {
	iFn
	params []IExpressionNode
	ExpressionNode
}

// AddParameter add node as current container's params
func (f *Fn) AddParameter(node IExpressionNode) {
	f.params = append(f.params, node)
	node.setContainer(f)
}

// Parameters return values of all parameter of this Container node.
// This is read-only, so we will return a slice of value
func (f *Fn) Parameters() []IExpressionNode {
	var result []IExpressionNode
	for _, p := range f.params {
		result = append(result, p)
	}
	return result
}

func (f *Fn) getLevel() (level int) {
	return f.level
}
