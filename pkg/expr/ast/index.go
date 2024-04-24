package ast

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expr/ast/base"
	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

type (
	index struct {
		base.Node
		base.Container
	}
)

func (i *index) Value() any {
	panic("not implemented")
}

func (i *index) Accept(eCtx interfaces.Context, v interfaces.Visitor) any {
	return v.VisitIndex(eCtx, i)
}

func (i *index) TraceFullyRealized() bool {
	return true
}

func (i *index) Expr() string {
	// Verify if we can simplify the expr, we would rather return
	// github.sha then github['sha'] so we check if this is a simple case.
	if lt, ok := i.Params[1].(*literal); ok {
		if lStr, ok := lt.Value().(string); ok && legalKeyWord(lStr) {
			return fmt.Sprintf("%s.%s", i.Params[0].Expr(), i.Params[0].Expr())
		}
	}
	return fmt.Sprintf("%s[%s]", i.Params[0].Expr(), i.Params[1].Expr())
}

func (i *index) RealizedExpr(eCtx interfaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(i)
	if exist {
		return result
	}
	return fmt.Sprintf("%s[%s]", i.Params[0].Expr(), i.Params[1].Expr())
}

func (i *index) GetCtn() interfaces.Container {
	return i.Ctn
}

func (i *index) SetCtn(c interfaces.Container) {
	i.Ctn = c
}

func (i *index) GetLevel() (level int) {
	return i.Level
}

func (i *index) GetName() string {
	return i.Name
}

func (i *index) SetName(name string) {
	i.Name = name
}
