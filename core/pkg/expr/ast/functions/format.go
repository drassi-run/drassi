package functions

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/dungdm93/drassi/core/pkg/expr/ast/ast_ifaces"
)

type Format struct {
	Fn
}

func (f *Format) Value() any {
	panic("not implemented")
}

func (f *Format) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitFormatFn(eCtx, f)
}

func (f *Format) TraceFullyRealized() bool {
	return false
}

func (f *Format) SetName(name string) {
	f.Name = name
}

func (f *Format) GetName() string {
	return f.Name
}

func (f *Format) GetCtn() ast_ifaces.Container {
	return f.Ctn
}

func (f *Format) SetCtn(c ast_ifaces.Container) {
	f.Ctn = c
}

func readArgIdx(str string, start int) (ok bool, argIndex int) {
	var length int
	if unicode.IsDigit(at(str, start+length)) {
		length++
	}
	if length < 1 {
		return false, -1
	}
	argIndex, err := strconv.Atoi(str[start : start+length])
	if err != nil {
		return false, -1
	}
	return true, argIndex
}

func at(str string, i int) rune {
	s := []rune(str)
	if i < 0 || i >= len(str) {
		return rune(0)
	}
	return s[i]
}

type formatResultBuilder struct {
	node     *Format
	ctx      ast_ifaces.Context
	segments []string
	cache    []*argValue
}

type argValue struct {
	stringResult string
}

func (a *argValue) StringResult() string {
	return a.stringResult
}

func (f *Format) Expr() string {
	params := make([]string, len(f.Parameters()))
	for i, param := range f.Parameters() {
		params[i] = param.Expr()
	}
	return fmt.Sprintf("%s(%s)", f.GetName(), strings.Join(params, ", "))
}

func (f *Format) RealizedExpr(eCtx ast_ifaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(f)
	if exist {
		return result
	}
	params := make([]string, len(f.Parameters()))
	for i, param := range f.Parameters() {
		params[i] = param.RealizedExpr(eCtx)
	}
	return fmt.Sprintf("%s(%s)", f.GetName(), strings.Join(params, ", "))
}
