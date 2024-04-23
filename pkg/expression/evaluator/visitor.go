package evaluator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression"
	"github.com/dungdm93/drasi/pkg/expression/parser"
	"github.com/dungdm93/drasi/pkg/runner"
)

var (
	ErrorsTemplateContextNotFound  = errors.New("template context not found")
	ErrorsExecutionContextNotFound = errors.New("execution context not found")
	ErrorsInvalidFormatArgIndex    = errors.New("invalid format argument index")
)

type expressionNodeVisitor struct {
}

func (e expressionNodeVisitor) VisitAlwaysFn(eCtx expression.IEvaluationContext, c expression.IExpressionNode) any {
	return true
}

func (e expressionNodeVisitor) VisitAnd(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	result := &EvaluationResult{}
	for _, param := range c.Parameters() {
		result = evaluateWithContext(eCtx, param)
		if result.IsFalsy() {
			return result.Value()
		}
	}
	return result.Value()
}

func (e expressionNodeVisitor) VisitCancelledFn(eCtx expression.IEvaluationContext, c expression.IExpressionNode) any {
	tplCtx := eCtx.State().(*runner.TemplateContext)
	if tplCtx == nil {
		panic(ErrorsTemplateContextNotFound)
	}
	// TODO: refactor me
	execCtx := tplCtx.State["IExecutionContext"].(runner.IExecutionContext)
	if execCtx == nil {
		panic(ErrorsExecutionContextNotFound)
	}
	return execCtx.JobContext().Status == runner.ActionResultCancelled
}

func (e expressionNodeVisitor) VisitContainsFn(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	l := evaluateWithContext(eCtx, c.Parameters()[0])
	if l.IsPrimitive() {
		lStr := l.ConvertToString()
		r := evaluateWithContext(eCtx, c.Parameters()[1])
		if r.IsPrimitive() {
			rStr := r.ConvertToString()
			return containsIgnoreCase(lStr, rStr)
		}
	}
	isCol, col := l.TryGetCollectionInterface()
	if isCol {
		if arr, isArr := col.(expression.IReadOnlyArray); isArr && arr.Count() > 0 {
			r := evaluateWithContext(eCtx, c.Parameters()[1])
			e := arr.Enumerator()
			for e.Next() {
				i := createIntermediateResult(eCtx, e.Value())
				if r.AbstractEqual(i) {
					return true
				}
			}
		}
	}
	return false
}

func (e expressionNodeVisitor) VisitContextValueNode(eCtx expression.IEvaluationContext, c expression.IExpressionNode) any {
	return eCtx.State().(*runner.TemplateContext).ExpressionValues[c.GetName()]
}

func (e expressionNodeVisitor) VisitEndsWithFn(eCtx expression.IEvaluationContext, c expression.IContainer) any {
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

func (e expressionNodeVisitor) VisitEqual(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	l := evaluateWithContext(eCtx, c.Parameters()[0])
	r := evaluateWithContext(eCtx, c.Parameters()[1])
	return l.AbstractEqual(r)
}

func (e expressionNodeVisitor) VisitFailureFn(eCtx expression.IEvaluationContext, c expression.IExpressionNode) any {
	tplCtx := eCtx.State().(*runner.TemplateContext)
	if tplCtx == nil {
		panic(ErrorsTemplateContextNotFound)
	}
	// TODO: refactor me
	execCtx := tplCtx.State["IExecutionContext"].(runner.IExecutionContext)
	if execCtx == nil {
		panic(ErrorsExecutionContextNotFound)
	}
	// Decide based on 'action_status' for composite MAIN steps and 'job.status' for pre, post and job-getLevel steps
	isComposite := execCtx.IsEmbedded() && execCtx.Stage() == runner.ActionRunStageMain
	if isComposite {
		// TODO: refactor me
		// If status is not parseable, evaluate actionStatus to ActionResultSuccess
		if actionStatus := execCtx.GetGitHubContext("action_status"); actionStatus != "" {
			if ok := runner.TryParseActionResult(actionStatus); ok {
				return actionStatus == fmt.Sprintf("%s", runner.ActionResultFailure)
			}
			return false
		}
	}
	return execCtx.JobContext().Status == runner.ActionResultFailure
}

func (e expressionNodeVisitor) VisitFormatFn(eCtx expression.IEvaluationContext, c expression.IContainer) any {
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
func (e expressionNodeVisitor) VisitFromJsonFn(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	json := evaluateWithContext(eCtx, c.Parameters()[0]).ConvertToString()
	// return runner.ToPipelineContextData(json)
	return json
}

func (e expressionNodeVisitor) VisitGreaterThan(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	l := evaluateWithContext(eCtx, c.Parameters()[0])
	r := evaluateWithContext(eCtx, c.Parameters()[1])
	return l.AbstractGreaterThan(r)
}

func (e expressionNodeVisitor) VisitGreaterThanOrEqual(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	l := evaluateWithContext(eCtx, c.Parameters()[0])
	r := evaluateWithContext(eCtx, c.Parameters()[1])
	return l.AbstractGreaterThan(r)
}

func (e expressionNodeVisitor) VisitIndex(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	l := evaluateWithContext(eCtx, c.Parameters()[0])
	isCol, col := l.TryGetCollectionInterface()
	if !isCol {
		_, isW := c.Parameters()[1].(*parser.WildCard)
		if isW {
			return newFilteredArray()
		}
		return nil
	}
	fa, isFilteredArray := col.(*filteredArray)
	if isFilteredArray {
		return handleFilteredArray(eCtx, fa, c)
	}
	obj, isObj := col.(expression.IReadOnlyObj)
	if isObj {
		return handleObject(eCtx, obj, c)
	}
	arr, isArr := col.(expression.IReadOnlyArray)
	if isArr {
		return handleArray(eCtx, arr, c)
	}
	return nil
}

func (e expressionNodeVisitor) VisitJoinFn(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	items := evaluateWithContext(eCtx, c.Parameters()[0])
	isCol, col := items.TryGetCollectionInterface()
	if isCol {
		arr, isArr := col.(expression.IReadOnlyArray)
		if isArr && arr.Count() > 0 {
			var result strings.Builder
			item := arr.GetValue(0)
			itemResult := createIntermediateResult(eCtx, item)
			itemStr := itemResult.ConvertToString()
			_, err := result.WriteString(itemStr)
			if err != nil {
				return ""
			}
			if arr.Count() > 1 {
				separator := ","
				if len(c.Parameters()) > 1 {
					separatorResult := evaluateWithContext(eCtx, c.Parameters()[1])
					if separatorResult.IsPrimitive() {
						separator = separatorResult.ConvertToString()
					}
				}
				e := arr.Enumerator()
				for e.Next() {
					result.WriteString(separator)
					nextItem := e.Value()
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

func (e expressionNodeVisitor) VisitLessThan(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	left := evaluateWithContext(eCtx, c.Parameters()[0])
	right := evaluateWithContext(eCtx, c.Parameters()[1])
	return left.AbstractLessThan(right)
}

func (e expressionNodeVisitor) VisitLessThanOrEqual(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	left := evaluateWithContext(eCtx, c.Parameters()[0])
	right := evaluateWithContext(eCtx, c.Parameters()[1])
	return left.AbstractLessThanOrEqual(right)
}

func (e expressionNodeVisitor) VisitLiteral(eCtx expression.IEvaluationContext, c expression.IExpressionNode) any {
	return c.Value()
}

func (e expressionNodeVisitor) VisitNoopFn(eCtx expression.IEvaluationContext, c expression.IExpressionNode) any {
	return nil
}

func (e expressionNodeVisitor) VisitNoopNamedValue(eCtx expression.IEvaluationContext, c expression.IExpressionNode) any {
	return nil
}

func (e expressionNodeVisitor) VisitNot(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	return evaluateWithContext(eCtx, c.Parameters()[0]).IsFalsy()
}

func (e expressionNodeVisitor) VisitNotEqual(eCtx expression.IEvaluationContext, c expression.IContainer) any {
	l := evaluateWithContext(eCtx, c.Parameters()[0])
	r := evaluateWithContext(eCtx, c.Parameters()[1])
	return l.AbstractNotEqual(r)
}

func (e expressionNodeVisitor) VisitOr(eCtx expression.IEvaluationContext, c expression.IContainer) any {
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

func (e expressionNodeVisitor) VisitStartsWithFn(eCtx expression.IEvaluationContext, c expression.IContainer) any {
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

func (e expressionNodeVisitor) VisitSuccessFn(eCtx expression.IEvaluationContext, c expression.IExpressionNode) any {
	tplCtx := eCtx.State().(*runner.TemplateContext)
	if tplCtx == nil {
		panic(ErrorsTemplateContextNotFound)
	}
	// TODO: refactor me
	execCtx := tplCtx.State["IExecutionContext"].(runner.IExecutionContext)
	if execCtx == nil {
		panic(ErrorsExecutionContextNotFound)
	}
	// Decide based on 'action_status' for composite MAIN steps and 'job.status' for pre, post and job-getLevel steps
	isComposite := execCtx.IsEmbedded() && execCtx.Stage() == runner.ActionRunStageMain
	if isComposite {
		// TODO: refactor me
		// If status is not parsable, evaluate actionStatus to ActionResultSuccess
		if actionStatus := execCtx.GetGitHubContext("action_status"); actionStatus != "" {
			if ok := runner.TryParseActionResult(actionStatus); ok {
				return actionStatus == fmt.Sprintf("%s", runner.ActionResultSuccess)
			}
			return true
		}
	}
	return execCtx.JobContext().Status == runner.ActionResultSuccess
}

func (e expressionNodeVisitor) VisitWildCard(eCtx expression.IEvaluationContext, c expression.IExpressionNode) any {
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
