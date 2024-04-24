package evaluator

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"

	"github.com/dungdm93/drasi/pkg/expr/ast"
	"github.com/dungdm93/drasi/pkg/expr/common"
	"github.com/dungdm93/drasi/pkg/expr/interfaces"
	"github.com/dungdm93/drasi/pkg/model/contexts"
)

type evaluationVisitor struct {
}

func (e evaluationVisitor) VisitAlwaysFn(eCtx interfaces.Context, c interfaces.Node) any {
	return true
}

func (e evaluationVisitor) VisitAnd(eCtx interfaces.Context, c interfaces.Container) any {
	result := &result{}
	for _, param := range c.Parameters() {
		result = evaluate(eCtx, param)
		if result.isFalsy() {
			return result.Value()
		}
	}
	return result.Value()
}

func (e evaluationVisitor) VisitCancelledFn(eCtx interfaces.Context, c interfaces.Node) any {
	tplCtx := eCtx.State().(*contexts.Expr)
	if tplCtx == nil {
		panic(ErrorExprContextNotFound)
	}
	execCtx := tplCtx.State
	if execCtx == nil {
		panic(ErrorExecutionContextNotFound)
	}
	return execCtx.Job.Status == contexts.ActionResultCancelled
}

func (e evaluationVisitor) VisitContainsFn(eCtx interfaces.Context, c interfaces.Container) any {
	l := evaluate(eCtx, c.Parameters()[0])
	if l.primitive() {
		lStr := l.string()
		r := evaluate(eCtx, c.Parameters()[1])
		if r.primitive() {
			rStr := r.string()
			return containsIgnoreCase(lStr, rStr)
		}
	}
	isCol, col := l.isCollection()
	if isCol {
		if arr, isArr := col.(common.Array); isArr && len(arr) > 0 {
			r := evaluate(eCtx, c.Parameters()[1])
			for _, value := range arr {
				i := createIntermediateResult(eCtx, value)
				if r.equal(i) {
					return true
				}
			}
		}
	}
	return false
}

func (e evaluationVisitor) VisitContextValueNode(eCtx interfaces.Context, c interfaces.Node) any {
	var target map[string]interface{}
	if err := mapstructure.Decode(*(eCtx.State().(*contexts.Expr).State), &target); err != nil {
		panic(err)
	}
	fmt.Printf("target: %+v\n", target)
	return target[c.GetName()]
}

func (e evaluationVisitor) VisitEndsWithFn(eCtx interfaces.Context, c interfaces.Container) any {
	l := evaluate(eCtx, c.Parameters()[0])
	if l.primitive() {
		lStr := l.string()
		r := evaluate(eCtx, c.Parameters()[1])
		if r.primitive() {
			rStr := r.string()
			return endsWithIgnoreCase(lStr, rStr)
		}
	}
	return false
}

func (e evaluationVisitor) VisitEqual(eCtx interfaces.Context, c interfaces.Container) any {
	l := evaluate(eCtx, c.Parameters()[0])
	r := evaluate(eCtx, c.Parameters()[1])
	return l.equal(r)
}

// TODO: composite action
// See https://github.com/dungdm93/drasi/blob/bfc21ce03ad75998a64d4a4718c7d648fea24f2a/pkg/expression/evaluator/visitor.go#L99
func (e evaluationVisitor) VisitFailureFn(eCtx interfaces.Context, c interfaces.Node) any {
	tplCtx := eCtx.State().(*contexts.Expr)
	if tplCtx == nil {
		panic(ErrorExprContextNotFound)
	}
	// TODO: refactor me
	execCtx := tplCtx.State
	return execCtx.Job.Status == contexts.ActionResultFailure
}

func (e evaluationVisitor) VisitFormatFn(eCtx interfaces.Context, c interfaces.Container) any {
	format := evaluate(eCtx, c.Parameters()[0]).string()
	var idx int
	result := newFormatResultBuilder(c, eCtx, len(format))
	for idx < len(format) {
		// Find first occurrence of opening brace and closing brace in sub-sequence from idx:len(format)-1
		lBraceIdx := strings.Index(format[idx:], "{") + idx
		rBraceIdx := strings.Index(format[idx:], "}") + idx
		// Left brace found
		if lBraceIdx >= idx && (rBraceIdx < idx || rBraceIdx > lBraceIdx) {
			// escaped left brace
			if at(format, lBraceIdx+1) == '{' {
				result.appendStatic(format[idx : lBraceIdx-idx+1])
				idx = lBraceIdx + 2
				continue
			}
			// Left brace, number, right brace
			if rBraceIdx > lBraceIdx+1 {
				if ok, argIdx := readArgIdx(format, lBraceIdx+1); ok {
					// Check parameter count
					if argIdx > len(c.Parameters())-2 {
						panic(ErrorInvalidFormatArgIndex)
					}
					// Append the portion before the left brace
					result.appendStatic(format[idx:lBraceIdx])
					// Append the arg
					result.appendArgument(argIdx)
					idx = rBraceIdx + 1
				} else {
					panic(fmt.Errorf("invalid format string %s", format))
				}
			}
			continue
		}
		// Only right brace found
		if rBraceIdx >= idx {
			// escaped right brace
			if at(format, rBraceIdx+1) == '}' {
				result.appendStatic(format[idx : rBraceIdx+1])
				idx = rBraceIdx + 2
			} else {
				panic(fmt.Errorf("invalid format string %s", format))
			}
			continue
		}
		// Last segment
		result.appendStatic(format[idx:])
	}
	return result.String()
}

// TODO: implement real logic with PipelineContextData
func (e evaluationVisitor) VisitFromJsonFn(eCtx interfaces.Context, c interfaces.Container) any {
	json := evaluate(eCtx, c.Parameters()[0]).string()
	// return runner.ToPipelineContextData(json)
	return json
}

func (e evaluationVisitor) VisitGreaterThan(eCtx interfaces.Context, c interfaces.Container) any {
	l := evaluate(eCtx, c.Parameters()[0])
	r := evaluate(eCtx, c.Parameters()[1])
	return l.greaterThan(r)
}

func (e evaluationVisitor) VisitGreaterThanOrEqual(eCtx interfaces.Context, c interfaces.Container) any {
	l := evaluate(eCtx, c.Parameters()[0])
	r := evaluate(eCtx, c.Parameters()[1])
	return l.greaterThan(r)
}

func (e evaluationVisitor) VisitIndex(eCtx interfaces.Context, c interfaces.Container) any {
	l := evaluate(eCtx, c.Parameters()[0])
	isCol, col := l.isCollection()
	if !isCol {
		_, isW := c.Parameters()[1].(*ast.WildCard)
		if isW {
			return newFilteredArray()
		}
		return nil
	}
	fa, isFilteredArray := col.(filteredArray)
	if isFilteredArray {
		return handleFilteredArray(eCtx, fa, c)
	}
	obj, isObj := col.(common.Obj)
	if isObj {
		return handleObject(eCtx, obj, c)
	}
	arr, isArr := col.(common.Array)
	if isArr {
		return handleArray(eCtx, arr, c)
	}
	return nil
}

func (e evaluationVisitor) VisitJoinFn(eCtx interfaces.Context, c interfaces.Container) any {
	items := evaluate(eCtx, c.Parameters()[0])
	isCol, col := items.isCollection()
	if isCol {
		arr, isArr := col.(common.Array)
		if isArr && len(arr) > 0 {
			var result strings.Builder
			item := arr[0]
			itemResult := createIntermediateResult(eCtx, item)
			itemStr := itemResult.string()
			_, err := result.WriteString(itemStr)
			if err != nil {
				return ""
			}
			if len(arr) > 1 {
				separator := ","
				if len(c.Parameters()) > 1 {
					separatorResult := evaluate(eCtx, c.Parameters()[1])
					if separatorResult.primitive() {
						separator = separatorResult.string()
					}
				}
				for _, value := range arr {
					result.WriteString(separator)
					nextItem := value
					nextItemResult := createIntermediateResult(eCtx, nextItem)
					nextItemStr := nextItemResult.string()
					result.WriteString(nextItemStr)
				}
			}
			return result.String()
		}
	}
	if items.primitive() {
		return items.string()
	}
	return ""
}

func (e evaluationVisitor) VisitLessThan(eCtx interfaces.Context, c interfaces.Container) any {
	left := evaluate(eCtx, c.Parameters()[0])
	right := evaluate(eCtx, c.Parameters()[1])
	return left.lessThan(right)
}

func (e evaluationVisitor) VisitLessThanOrEqual(eCtx interfaces.Context, c interfaces.Container) any {
	left := evaluate(eCtx, c.Parameters()[0])
	right := evaluate(eCtx, c.Parameters()[1])
	return left.lessThanOrEqual(right)
}

func (e evaluationVisitor) VisitLiteral(eCtx interfaces.Context, c interfaces.Node) any {
	return c.Value()
}

func (e evaluationVisitor) VisitNoopFn(eCtx interfaces.Context, c interfaces.Node) any {
	return nil
}

func (e evaluationVisitor) VisitNoopNamedValue(eCtx interfaces.Context, c interfaces.Node) any {
	return nil
}

func (e evaluationVisitor) VisitNot(eCtx interfaces.Context, c interfaces.Container) any {
	return evaluate(eCtx, c.Parameters()[0]).isFalsy()
}

func (e evaluationVisitor) VisitNotEqual(eCtx interfaces.Context, c interfaces.Container) any {
	l := evaluate(eCtx, c.Parameters()[0])
	r := evaluate(eCtx, c.Parameters()[1])
	return l.notEqual(r)
}

func (e evaluationVisitor) VisitOr(eCtx interfaces.Context, c interfaces.Container) any {
	var result *result
	for _, p := range c.Parameters() {
		result = evaluate(eCtx, p)
		if result.isTruthy() {
			break
		}
	}
	if result == nil {
		return nil
	}
	return result.Value()
}

func (e evaluationVisitor) VisitStartsWithFn(eCtx interfaces.Context, c interfaces.Container) any {
	l := evaluate(eCtx, c.Parameters()[0])
	if l.primitive() {
		lStr := l.string()
		r := evaluate(eCtx, c.Parameters()[1])
		if r.primitive() {
			rStr := r.string()
			return startsWithIgnoreCase(lStr, rStr)
		}
	}
	return false
}

// TODO: composite action
// See https://github.com/dungdm93/drasi/blob/bfc21ce03ad75998a64d4a4718c7d648fea24f2a/pkg/expression/evaluator/visitor.go#L320
func (e evaluationVisitor) VisitSuccessFn(eCtx interfaces.Context, c interfaces.Node) any {
	tplCtx := eCtx.State().(*contexts.Expr)
	if tplCtx == nil {
		panic(ErrorExprContextNotFound)
	}
	ctx := tplCtx.State
	return ctx.Job.Status == contexts.ActionResultSuccess
}

func (e evaluationVisitor) VisitWildCard(eCtx interfaces.Context, c interfaces.Node) any {
	return common.Wildcard
}

func containsIgnoreCase(s string, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}

func endsWithIgnoreCase(str string, suffix string) bool {
	return strings.HasSuffix(strings.ToLower(str), strings.ToLower(suffix))
}

func startsWithIgnoreCase(str string, suffix string) bool {
	return strings.HasPrefix(strings.ToLower(str), strings.ToLower(suffix))
}
