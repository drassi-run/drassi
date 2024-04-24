package ast

import (
	"github.com/dungdm93/drasi/pkg/expr/ast/base"
	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

type INamedValue interface {
	interfaces.Node
}

// NamedValue represent keyword available from parseContext eg: github, job, steps, env
type NamedValue struct {
	INamedValue
	base.Node
}

func (n *NamedValue) Expr() string {
	return n.Name
}

func (n *NamedValue) RealizedExpr(eCtx interfaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(n)
	if exist {
		return result
	}
	return n.Name
}

func (n *NamedValue) GetLevel() (level int) {
	return n.Level
}
