package ast

import (
	"fmt"

	"drassi.run/core/pkg/expression/ast/ast_ifaces"
	"drassi.run/core/pkg/expression/ast/base"
	"drassi.run/core/pkg/expression/token"
)

type (
	Index struct {
		base.Node
		base.Container
	}
)

func (i *Index) Value() any {
	panic("not implemented")
}

func (i *Index) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitIndex(eCtx, i)
}

func (i *Index) TraceFullyRealized() bool {
	return true
}

func (i *Index) Expr() string {
	// Verify if we can simplify the expr, we would rather return
	// github.sha then github['sha'] so we check if this is a simple case.
	if lt, ok := i.Params[1].(*Lit); ok {
		if lStr, ok := lt.Value().(string); ok && token.LegalKeyWord(lStr) {
			return fmt.Sprintf("%s.%s", i.Params[0].Expr(), i.Params[0].Expr())
		}
	}
	return fmt.Sprintf("%s[%s]", i.Params[0].Expr(), i.Params[1].Expr())
}

func (i *Index) RealizedExpr(eCtx ast_ifaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(i)
	if exist {
		return result
	}
	return fmt.Sprintf("%s[%s]", i.Params[0].Expr(), i.Params[1].Expr())
}

func (i *Index) GetCtn() ast_ifaces.Container {
	return i.Ctn
}

func (i *Index) SetCtn(c ast_ifaces.Container) {
	i.Ctn = c
}

func (i *Index) GetLevel() (level int) {
	return i.Level
}

func (i *Index) GetName() string {
	return i.Name
}

func (i *Index) SetName(name string) {
	i.Name = name
}
