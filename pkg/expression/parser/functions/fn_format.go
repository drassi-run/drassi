package functions

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/dungdm93/drasi/pkg/expression"
)

var (
	ErrorsInvalidFormatArgIndex = errors.New("invalid format argument index")
)

type FormatFn struct {
	Fn
}

func (f *FormatFn) Value() any {
	panic("not implemented")
}

func (f *FormatFn) Accept(eCtx expression.IEvaluationContext, v expression.IExpressionNodeVisitor) any {
	return v.VisitFormatFn(eCtx, f)
}

func (f *FormatFn) TraceFullyRealized() bool {
	return false
}

func (f *FormatFn) SetName(name string) {
	f.Name = name
}

func (f *FormatFn) GetName() string {
	return f.Name
}

func (f *FormatFn) GetContainer() expression.IContainer {
	return f.Container
}

func (f *FormatFn) SetContainer(c expression.IContainer) {
	f.Container = c
}

//
// func (f *FormatFn) EvaluateCore(eCtx interfaces.IEvaluationContext) any {
// 	format := evaluator.EvaluateWithContext(eCtx, f.Parameters()[0]).ConvertToString()
// 	var idx int
// 	result := newFormatResultBuilder(f, eCtx, len(format))
// 	for idx < len(format) {
// 		// Find first occurrence of opening brace and closing brace in sub-sequence from idx:len(format)-1
// 		lBraceIdx := strings.Index(format[idx:], "{") + idx
// 		rBraceIdx := strings.Index(format[idx:], "}") + idx
// 		// Left brace found
// 		if lBraceIdx >= idx && (rBraceIdx < idx || rBraceIdx > lBraceIdx) {
// 			// escaped left brace
// 			if at(format, lBraceIdx+1) == '{' {
// 				result.appendStatic(format[idx : lBraceIdx-idx+1])
// 				idx = lBraceIdx + 2
// 				continue
// 			}
// 			// Left brace, number, right brace
// 			if rBraceIdx > lBraceIdx+1 {
// 				if ok, argIdx := readArgIdx(format, lBraceIdx+1); ok {
// 					// Check parameter count
// 					if argIdx > len(f.Parameters())-2 {
// 						panic(ErrorsInvalidFormatArgIndex)
// 					}
// 					// Append the portion before the left brace
// 					result.appendStatic(format[idx:lBraceIdx])
// 					// Append the arg
// 					result.appendArgument(argIdx)
// 					idx = rBraceIdx + 1
// 				} else {
// 					panic(fmt.Errorf("invalid format string %s", format))
// 				}
// 			}
// 			continue
// 		}
// 		// Only right brace found
// 		if rBraceIdx >= idx {
// 			// escaped right brace
// 			if at(format, rBraceIdx+1) == '}' {
// 				result.appendStatic(format[idx : rBraceIdx+1])
// 				idx = rBraceIdx + 2
// 			} else {
// 				panic(fmt.Errorf("invalid format string %s", format))
// 			}
// 			continue
// 		}
// 		// Last segment
// 		result.appendStatic(format[idx:])
// 	}
// 	return result.ValueKindString()
// }

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
	node     *FormatFn
	ctx      expression.IEvaluationContext
	segments []string
	cache    []*argValue
}

type argValue struct {
	stringResult string
}

func (a *argValue) StringResult() string {
	return a.stringResult
}

//
// func newFormatResultBuilder(node *FormatFn, ctx interfaces.IEvaluationContext, cacheSize int) *formatResultBuilder {
// 	return &formatResultBuilder{node: node, ctx: ctx, segments: []string{}, cache: make([]*argValue, cacheSize)}
// }
//
// func (f *formatResultBuilder) ValueKindString() string {
// 	return strings.Join(f.segments, "")
// }
//
// func (f *formatResultBuilder) appendStatic(value string) {
// 	if len(value) > 0 {
// 		f.segments = append(f.segments, value)
// 	}
// }
//
// func (f *formatResultBuilder) appendArgument(argIdx int) {
// 	var result string
// 	argVal := f.cache[argIdx]
// 	if argVal == nil {
// 		// cache miss
// 		var evaluationResult = evaluator.EvaluateWithContext(f.ctx, f.node.Parameters()[argIdx+1])
// 		argVal = &argValue{
// 			stringResult: evaluationResult.ConvertToString(),
// 		}
// 		f.cache = append(f.cache, argVal)
// 	}
// 	result = argVal.StringResult()
// 	f.appendStatic(result)
// }

func (f *FormatFn) ConvertToExpression() string {
	params := make([]string, len(f.Parameters()))
	for i, param := range f.Parameters() {
		params[i] = param.ConvertToExpression()
	}
	return fmt.Sprintf("%s(%s)", f.GetName(), strings.Join(params, ", "))
}

func (f *FormatFn) ConvertToRealizedExpression(eCtx expression.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(f)
	if exist {
		return result
	}
	params := make([]string, len(f.Parameters()))
	for i, param := range f.Parameters() {
		params[i] = param.ConvertToRealizedExpression(eCtx)
	}
	return fmt.Sprintf("%s(%s)", f.GetName(), strings.Join(params, ", "))
}
