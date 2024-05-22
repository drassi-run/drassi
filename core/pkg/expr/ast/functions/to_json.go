package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drassi/core/pkg/expr/ast/ast_ifaces"
)

type ToJson struct {
	Fn
}


func (t *ToJson) Value() any {
	panic("not implemented")
}

func (t *ToJson) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitToJsonFn(eCtx, t)
}

func (t *ToJson) TraceFullyRealized() bool {
	return false
}

func (t *ToJson) SetName(name string) {
	t.Name = name
}

func (t *ToJson) GetName() string {
	return t.Name
}

func (t *ToJson) GetCtn() ast_ifaces.Container {
	return t.Ctn
}

func (t *ToJson) SetCtn(c ast_ifaces.Container) {
	t.Ctn = c
}

func (t *ToJson) Expr() string {
	params := make([]string, len(t.Parameters()))
	for i, param := range t.Parameters() {
		params[i] = param.Expr()
	}
	return fmt.Sprintf("%s(%s)", t.GetName(), strings.Join(params, ", "))
}

func (t *ToJson) RealizedExpr(eCtx ast_ifaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(t)
	if exist {
		return result
	}
	params := make([]string, len(t.Parameters()))
	for i, param := range t.Parameters() {
		params[i] = param.RealizedExpr(eCtx)
	}
	return fmt.Sprintf("%s(%s)", t.GetName(), strings.Join(params, ", "))
}
