package operators

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drassi/core/pkg/expression/ast/ast_ifaces"
	"github.com/dungdm93/drassi/core/pkg/expression/ast/base"
)

type And struct {
	base.Container
	base.Node
}

func (a *And) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitAnd(eCtx, a)
}

func (a *And) Value() any {
	panic("not implemented")
}

func (a *And) Expr() string {
	expressions := make([]string, len(a.Params))
	for i, param := range a.Params {
		expressions[i] = param.Expr()
	}
	return fmt.Sprintf("(%s)", strings.Join(expressions, " && "))
}

func (a *And) RealizedExpr(eCtx ast_ifaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(a)
	if exist {
		return result
	}
	expressions := make([]string, len(a.Params))
	for i, param := range a.Params {
		expressions[i] = param.RealizedExpr(eCtx)
	}
	return fmt.Sprintf("(%s)", strings.Join(expressions, " && "))
}

func (a *And) TraceFullyRealized() bool {
	return false
}

func (a *And) GetCtn() ast_ifaces.Container {
	return a.Ctn
}

func (a *And) SetCtn(c ast_ifaces.Container) {
	a.Ctn = c
}

func (a *And) GetLevel() (level int) {
	return a.Level
}

func (a *And) GetName() string {
	return a.Name
}

func (a *And) SetName(name string) {
	a.Name = name
}
