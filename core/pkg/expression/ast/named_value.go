package ast

import (
	"drassi.run/core/pkg/expression/ast/ast_ifaces"
	"drassi.run/core/pkg/expression/ast/base"
)

type INamedValue interface {
	ast_ifaces.ExprNode
}

// NamedValue represent keyword available from parseContext eg: github, job, steps, env
type NamedValue struct {
	INamedValue
	base.Node
}

func (n *NamedValue) Expr() string {
	return n.Name
}

func (n *NamedValue) RealizedExpr(eCtx ast_ifaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(n)
	if exist {
		return result
	}
	return n.Name
}

func (n *NamedValue) GetLevel() (level int) {
	return n.Level
}
