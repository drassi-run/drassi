package evaluator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression"
	"github.com/dungdm93/drasi/pkg/expression/parser"
	"github.com/dungdm93/drasi/pkg/model/contexts"
	"github.com/dungdm93/drasi/pkg/runner"
)

var (
	ErrorsTemplateContextNotFound  = errors.New("template context not found")
	ErrorsExecutionContextNotFound = errors.New("execution context not found")
	ErrorsInvalidFormatArgIndex    = errors.New("invalid format argument index")
)

type expNodeVisitor struct {
}

func (e expNodeVisitor) VisitAlwaysFn(eCtx expression.IEvaluationContext, c expression.IExpNode) any {
	return true
}

func (e expNodeVisitor) VisitAnd(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	result := &EvaluationResult{}
	for _, param := range c.Parameters() {
		result = evaluateWithContext(eCtx, param)
		if result.IsFalsy() {
			return result.Value()
		}
	}
	return result.Value()
}

func (e expNodeVisitor) VisitCancelledFn(eCtx expression.IEvaluationContext, c expression.IExpNode) any {
	tplCtx := eCtx.State().(*runner.TemplateContext)
	if tplCtx == nil {
		panic(ErrorsTemplateContextNotFound)
	}
	// TODO: refactor me
	execCtx := tplCtx.State["IExecutionContext"].(*contexts.Context)
	if execCtx == nil {
		panic(ErrorsExecutionContextNotFound)
	}
	return execCtx.Job.Status == contexts.ActionResultCancelled
}

func (e expNodeVisitor) VisitContainsFn(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	l := evaluateWithContext(eCtx, c.Parameters()[0])
	if l.IsPrimitive() {
		lStr := l.ConvertToString()
		r := evaluateWithContext(eCtx, c.Parameters()[1])
		if r.IsPrimitive() {
			rStr := r.ConvertToString()
			return containsIgnoreCase(lStr, rStr)
		}
	}
	isCol, col := l.IsCollection()
	if isCol {
		if arr, isArr := col.(expression.ReadOnlyArray); isArr && len(arr) > 0 {
			r := evaluateWithContext(eCtx, c.Parameters()[1])
			for _, value := range arr {
				i := createIntermediateResult(eCtx, value)
				if r.AbstractEqual(i) {
					return true
				}
			}
		}
	}
	return false
}

func (e expNodeVisitor) VisitContextValueNode(eCtx expression.IEvaluationContext, c expression.IExpNode) any {
	return eCtx.State().(*runner.TemplateContext).ExpressionValues[c.GetName()]
}

func (e expNodeVisitor) VisitEndsWithFn(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	l := evaluateWithContext(eCtx, c.Parameters()[0])
	if l.IsPrimitive() {
		lStr := l.ConvertToString()
		r := evaluateWithContext(eCtx, c.Parameters()[1])
		if r.IsPrimitive() {
			rStr := r.ConvertToString()
			return endsWithIgnoreCase(lStr, rStr)
		}
	}
	return false
}

func (e expNodeVisitor) VisitEqual(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	l := evaluateWithContext(eCtx, c.Parameters()[0])
	r := evaluateWithContext(eCtx, c.Parameters()[1])
	return l.AbstractEqual(r)
}

// TODO: composite action
// See https://github.com/dungdm93/drasi/blob/bfc21ce03ad75998a64d4a4718c7d648fea24f2a/pkg/expression/evaluator/visitor.go#L99
func (e expNodeVisitor) VisitFailureFn(eCtx expression.IEvaluationContext, c expression.IExpNode) any {
	tplCtx := eCtx.State().(*runner.TemplateContext)
	if tplCtx == nil {
		panic(ErrorsTemplateContextNotFound)
	}
	// TODO: refactor me
	execCtx := tplCtx.State["IExecutionContext"].(*contexts.Context)
	return execCtx.Job.Status == contexts.ActionResultFailure
}

func (e expNodeVisitor) VisitFormatFn(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	format := evaluateWithContext(eCtx, c.Parameters()[0]).ConvertToString()
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
						panic(ErrorsInvalidFormatArgIndex)
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
func (e expNodeVisitor) VisitFromJsonFn(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	json := evaluateWithContext(eCtx, c.Parameters()[0]).ConvertToString()
	// return runner.ToPipelineContextData(json)
	return json
}

func (e expNodeVisitor) VisitGreaterThan(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	l := evaluateWithContext(eCtx, c.Parameters()[0])
	r := evaluateWithContext(eCtx, c.Parameters()[1])
	return l.AbstractGreaterThan(r)
}

func (e expNodeVisitor) VisitGreaterThanOrEqual(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	l := evaluateWithContext(eCtx, c.Parameters()[0])
	r := evaluateWithContext(eCtx, c.Parameters()[1])
	return l.AbstractGreaterThan(r)
}

func (e expNodeVisitor) VisitIndex(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	l := evaluateWithContext(eCtx, c.Parameters()[0])
	isCol, col := l.IsCollection()
	if !isCol {
		_, isW := c.Parameters()[1].(*parser.WildCard)
		if isW {
			return newFilteredArray()
		}
		return nil
	}
	fa, isFilteredArray := col.(filteredArray)
	if isFilteredArray {
		return handleFilteredArray(eCtx, fa, c)
	}
	obj, isObj := col.(expression.ReadOnlyObj)
	if isObj {
		return handleObject(eCtx, obj, c)
	}
	arr, isArr := col.(expression.ReadOnlyArray)
	if isArr {
		return handleArray(eCtx, arr, c)
	}
	return nil
}

func (e expNodeVisitor) VisitJoinFn(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	items := evaluateWithContext(eCtx, c.Parameters()[0])
	isCol, col := items.IsCollection()
	if isCol {
		arr, isArr := col.(expression.ReadOnlyArray)
		if isArr && len(arr) > 0 {
			var result strings.Builder
			item := arr[0]
			itemResult := createIntermediateResult(eCtx, item)
			itemStr := itemResult.ConvertToString()
			_, err := result.WriteString(itemStr)
			if err != nil {
				return ""
			}
			if len(arr) > 1 {
				separator := ","
				if len(c.Parameters()) > 1 {
					separatorResult := evaluateWithContext(eCtx, c.Parameters()[1])
					if separatorResult.IsPrimitive() {
						separator = separatorResult.ConvertToString()
					}
				}
				for _, value := range arr {
					result.WriteString(separator)
					nextItem := value
					nextItemResult := createIntermediateResult(eCtx, nextItem)
					nextItemStr := nextItemResult.ConvertToString()
					result.WriteString(nextItemStr)
				}
			}
			return result.String()
		}
	}
	if items.IsPrimitive() {
		return items.ConvertToString()
	}
	return ""
}

func (e expNodeVisitor) VisitLessThan(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	left := evaluateWithContext(eCtx, c.Parameters()[0])
	right := evaluateWithContext(eCtx, c.Parameters()[1])
	return left.AbstractLessThan(right)
}

func (e expNodeVisitor) VisitLessThanOrEqual(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	left := evaluateWithContext(eCtx, c.Parameters()[0])
	right := evaluateWithContext(eCtx, c.Parameters()[1])
	return left.AbstractLessThanOrEqual(right)
}

func (e expNodeVisitor) VisitLiteral(eCtx expression.IEvaluationContext, c expression.IExpNode) any {
	return c.Value()
}

func (e expNodeVisitor) VisitNoopFn(eCtx expression.IEvaluationContext, c expression.IExpNode) any {
	return nil
}

func (e expNodeVisitor) VisitNoopNamedValue(eCtx expression.IEvaluationContext, c expression.IExpNode) any {
	return nil
}

func (e expNodeVisitor) VisitNot(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	return evaluateWithContext(eCtx, c.Parameters()[0]).IsFalsy()
}

func (e expNodeVisitor) VisitNotEqual(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	l := evaluateWithContext(eCtx, c.Parameters()[0])
	r := evaluateWithContext(eCtx, c.Parameters()[1])
	return l.AbstractNotEqual(r)
}

func (e expNodeVisitor) VisitOr(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	var result *EvaluationResult
	for _, p := range c.Parameters() {
		result = evaluateWithContext(eCtx, p)
		if result.IsTruthy() {
			break
		}
	}
	if result == nil {
		return nil
	}
	return result.Value()
}

func (e expNodeVisitor) VisitStartsWithFn(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	l := evaluateWithContext(eCtx, c.Parameters()[0])
	if l.IsPrimitive() {
		lStr := l.ConvertToString()
		r := evaluateWithContext(eCtx, c.Parameters()[1])
		if r.IsPrimitive() {
			rStr := r.ConvertToString()
			return startsWithIgnoreCase(lStr, rStr)
		}
	}
	return false
}

// TODO: composite action
// See https://github.com/dungdm93/drasi/blob/bfc21ce03ad75998a64d4a4718c7d648fea24f2a/pkg/expression/evaluator/visitor.go#L320
func (e expNodeVisitor) VisitSuccessFn(eCtx expression.IEvaluationContext, c expression.IExpNode) any {
	tplCtx := eCtx.State().(*runner.TemplateContext)
	if tplCtx == nil {
		panic(ErrorsTemplateContextNotFound)
	}
	ctx := tplCtx.State["IExecutionContext"].(*contexts.Context)
	return ctx.Job.Status == contexts.ActionResultSuccess
}

func (e expNodeVisitor) VisitWildCard(eCtx expression.IEvaluationContext, c expression.IExpNode) any {
	return expression.Wildcard
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
