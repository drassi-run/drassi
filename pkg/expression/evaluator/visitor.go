package evaluator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression/constants"
	"github.com/dungdm93/drasi/pkg/expression/interfaces"
	"github.com/dungdm93/drasi/pkg/expression/parser"
	"github.com/dungdm93/drasi/pkg/runner"
)

var (
	ErrorsTemplateContextNotFound  = errors.New("template context not found")
	ErrorsExecutionContextNotFound = errors.New("execution context not found")
	ErrorsInvalidFormatArgIndex    = errors.New("invalid format argument index")
)

type ExpressionNodeVisitor struct {
}

func (e ExpressionNodeVisitor) VisitAlwaysFn(eCtx interfaces.IEvaluationContext, c interfaces.IExpressionNode) any {
	return true
}

func (e ExpressionNodeVisitor) VisitAnd(eCtx interfaces.IEvaluationContext, c interfaces.IContainer) any {
	result := &EvaluationResult{}
	for _, param := range c.Parameters() {
		result = EvaluateWithContext(eCtx, param)
		if result.IsFalsy() {
			return result.Value()
		}
	}
	return result.Value()
}

func (e ExpressionNodeVisitor) VisitCancelledFn(eCtx interfaces.IEvaluationContext, c interfaces.IExpressionNode) any {
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

func (e ExpressionNodeVisitor) VisitContainsFn(eCtx interfaces.IEvaluationContext, c interfaces.IContainer) any {
	l := EvaluateWithContext(eCtx, c.Parameters()[0])
	if l.IsPrimitive() {
		lStr := l.ConvertToString()
		r := EvaluateWithContext(eCtx, c.Parameters()[1])
		if r.IsPrimitive() {
			rStr := r.ConvertToString()
			return containsIgnoreCase(lStr, rStr)
		}
	}
	isCol, col := l.TryGetCollectionInterface()
	if isCol {
		if arr, isArr := col.(interfaces.IReadOnlyArray); isArr && arr.Count() > 0 {
			r := EvaluateWithContext(eCtx, c.Parameters()[1])
			e := arr.Enumerator()
			for e.Next() {
				i := CreateIntermediateResult(eCtx, e.Value())
				if r.AbstractEqual(i) {
					return true
				}
			}
		}
	}
	return false
}

func (e ExpressionNodeVisitor) VisitContextValueNode(eCtx interfaces.IEvaluationContext, c interfaces.IExpressionNode) any {
	return eCtx.State().(*runner.TemplateContext).ExpressionValues[c.GetName()]
}

func (e ExpressionNodeVisitor) VisitEndsWithFn(eCtx interfaces.IEvaluationContext, c interfaces.IContainer) any {
	l := EvaluateWithContext(eCtx, c.Parameters()[0])
	if l.IsPrimitive() {
		lStr := l.ConvertToString()
		r := EvaluateWithContext(eCtx, c.Parameters()[1])
		if r.IsPrimitive() {
			rStr := r.ConvertToString()
			return endsWithIgnoreCase(lStr, rStr)
		}
	}
	return false
}

func (e ExpressionNodeVisitor) VisitEqual(eCtx interfaces.IEvaluationContext, c interfaces.IContainer) any {
	l := EvaluateWithContext(eCtx, c.Parameters()[0])
	r := EvaluateWithContext(eCtx, c.Parameters()[1])
	return l.AbstractEqual(r)
}

func (e ExpressionNodeVisitor) VisitFailureFn(eCtx interfaces.IEvaluationContext, c interfaces.IExpressionNode) any {
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

func (e ExpressionNodeVisitor) VisitFormatFn(eCtx interfaces.IEvaluationContext, c interfaces.IContainer) any {
	format := EvaluateWithContext(eCtx, c.Parameters()[0]).ConvertToString()
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
func (e ExpressionNodeVisitor) VisitFromJsonFn(eCtx interfaces.IEvaluationContext, c interfaces.IContainer) any {
	json := EvaluateWithContext(eCtx, c.Parameters()[0]).ConvertToString()
	// return runner.ToPipelineContextData(json)
	return json
}

func (e ExpressionNodeVisitor) VisitGreaterThan(eCtx interfaces.IEvaluationContext, c interfaces.IContainer) any {
	l := EvaluateWithContext(eCtx, c.Parameters()[0])
	r := EvaluateWithContext(eCtx, c.Parameters()[1])
	return l.AbstractGreaterThan(r)
}

func (e ExpressionNodeVisitor) VisitGreaterThanOrEqual(eCtx interfaces.IEvaluationContext, c interfaces.IContainer) any {
	l := EvaluateWithContext(eCtx, c.Parameters()[0])
	r := EvaluateWithContext(eCtx, c.Parameters()[1])
	return l.AbstractGreaterThan(r)
}

func (e ExpressionNodeVisitor) VisitIndex(eCtx interfaces.IEvaluationContext, c interfaces.IContainer) any {
	l := EvaluateWithContext(eCtx, c.Parameters()[0])
	isCol, col := l.TryGetCollectionInterface()
	if !isCol {
		_, isW := c.Parameters()[1].(*parser.WildCard)
		if isW {
			return newFilteredArray()
		}
		return nil
	}
	fa, isFilteredArray := col.(*FilteredArray)
	if isFilteredArray {
		return handleFilteredArray(eCtx, fa, c)
	}
	obj, isObj := col.(interfaces.IReadOnlyObj)
	if isObj {
		return handleObject(eCtx, obj, c)
	}
	arr, isArr := col.(interfaces.IReadOnlyArray)
	if isArr {
		return handleArray(eCtx, arr, c)
	}
	return nil
}

func (e ExpressionNodeVisitor) VisitJoinFn(eCtx interfaces.IEvaluationContext, c interfaces.IContainer) any {
	items := EvaluateWithContext(eCtx, c.Parameters()[0])
	isCol, col := items.TryGetCollectionInterface()
	if isCol {
		arr, isArr := col.(interfaces.IReadOnlyArray)
		if isArr && arr.Count() > 0 {
			var result strings.Builder
			item := arr.GetValue(0)
			itemResult := CreateIntermediateResult(eCtx, item)
			itemStr := itemResult.ConvertToString()
			_, err := result.WriteString(itemStr)
			if err != nil {
				return ""
			}
			if arr.Count() > 1 {
				separator := ","
				if len(c.Parameters()) > 1 {
					separatorResult := EvaluateWithContext(eCtx, c.Parameters()[1])
					if separatorResult.IsPrimitive() {
						separator = separatorResult.ConvertToString()
					}
				}
				e := arr.Enumerator()
				for e.Next() {
					result.WriteString(separator)
					nextItem := e.Value()
					nextItemResult := CreateIntermediateResult(eCtx, nextItem)
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

func (e ExpressionNodeVisitor) VisitLessThan(eCtx interfaces.IEvaluationContext, c interfaces.IContainer) any {
	left := EvaluateWithContext(eCtx, c.Parameters()[0])
	right := EvaluateWithContext(eCtx, c.Parameters()[1])
	return left.AbstractLessThan(right)
}

func (e ExpressionNodeVisitor) VisitLessThanOrEqual(eCtx interfaces.IEvaluationContext, c interfaces.IContainer) any {
	left := EvaluateWithContext(eCtx, c.Parameters()[0])
	right := EvaluateWithContext(eCtx, c.Parameters()[1])
	return left.AbstractLessThanOrEqual(right)
}

func (e ExpressionNodeVisitor) VisitLiteral(eCtx interfaces.IEvaluationContext, c interfaces.IExpressionNode) any {
	return c.Value()
}

func (e ExpressionNodeVisitor) VisitNoopFn(eCtx interfaces.IEvaluationContext, c interfaces.IExpressionNode) any {
	return nil
}

func (e ExpressionNodeVisitor) VisitNoopNamedValue(eCtx interfaces.IEvaluationContext, c interfaces.IExpressionNode) any {
	return nil
}

func (e ExpressionNodeVisitor) VisitNot(eCtx interfaces.IEvaluationContext, c interfaces.IContainer) any {
	return EvaluateWithContext(eCtx, c.Parameters()[0]).IsFalsy()
}

func (e ExpressionNodeVisitor) VisitNotEqual(eCtx interfaces.IEvaluationContext, c interfaces.IContainer) any {
	l := EvaluateWithContext(eCtx, c.Parameters()[0])
	r := EvaluateWithContext(eCtx, c.Parameters()[1])
	return l.AbstractNotEqual(r)
}

func (e ExpressionNodeVisitor) VisitOr(eCtx interfaces.IEvaluationContext, c interfaces.IContainer) any {
	var result *EvaluationResult
	for _, p := range c.Parameters() {
		result = EvaluateWithContext(eCtx, p)
		if result.IsTruthy() {
			break
		}
	}
	if result == nil {
		return nil
	}
	return result.Value()
}

func (e ExpressionNodeVisitor) VisitStartsWithFn(eCtx interfaces.IEvaluationContext, c interfaces.IContainer) any {
	l := EvaluateWithContext(eCtx, c.Parameters()[0])
	if l.IsPrimitive() {
		lStr := l.ConvertToString()
		r := EvaluateWithContext(eCtx, c.Parameters()[1])
		if r.IsPrimitive() {
			rStr := r.ConvertToString()
			return startsWithIgnoreCase(lStr, rStr)
		}
	}
	return false
}

func (e ExpressionNodeVisitor) VisitSuccessFn(eCtx interfaces.IEvaluationContext, c interfaces.IExpressionNode) any {
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

func (e ExpressionNodeVisitor) VisitWildCard(eCtx interfaces.IEvaluationContext, c interfaces.IExpressionNode) any {
	return constants.Wildcard
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
