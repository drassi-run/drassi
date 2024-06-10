package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drassi/core/pkg/expression/ast/ast_ifaces"
)

type Join struct {
	Fn
}

func (j *Join) Value() any {
	panic("not implemented")
}

// Generally this should not be overridden. True indicates the result of the node is traced as part of the "expanded"
// (i.e. "realized") trace information. Otherwise, the node expression is printed, and parameters to the node may or
// may not be fully realized - depending on each respective parameter's trace-fully-realized setting.
//
// The purpose is so the end user can understand how their expression expanded at run time. For example, consider
// the expression: eq(variables.publish, 'true'). The runtime-expanded expression may be: eq('true', 'true')
func (j *Join) TraceFullyRealized() bool {
	return true
}

func (j *Join) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitJoinFn(eCtx, j)
}

func (j *Join) SetName(name string) {
	j.Name = name
}

func (j *Join) GetName() string {
	return j.Name
}

func (j *Join) GetCtn() ast_ifaces.Container {
	return j.Ctn
}

func (j *Join) SetCtn(c ast_ifaces.Container) {
	j.Ctn = c
}

func (j *Join) Expr() string {
	params := make([]string, len(j.Parameters()))
	for i, param := range j.Parameters() {
		params[i] = param.Expr()
	}
	return fmt.Sprintf("%s(%s)", j.GetName(), strings.Join(params, ", "))
}

func (j *Join) RealizedExpr(eCtx ast_ifaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(j)
	if exist {
		return result
	}
	params := make([]string, len(j.Parameters()))
	for i, param := range j.Parameters() {
		params[i] = param.RealizedExpr(eCtx)
	}
	return fmt.Sprintf("%s(%s)", j.GetName(), strings.Join(params, ", "))
}
