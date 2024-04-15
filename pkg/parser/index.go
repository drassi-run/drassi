package parser

import (
	"fmt"
)

type (
	Index struct {
		ExpressionNode
		Container
	}

	IndexHelper struct {
		param  IExpressionNode
		result *EvaluationResult
		intIdx *int
		strIdx *string
	}
)

func (i *Index) traceFullyRealized() bool {
	return true
}

func (i *Index) convertToExpression() string {
	// Verify if we can simplify the expression, we would rather return
	// github.sha then github['sha'] so we check if this is a simple case.
	if lt, ok := i.params[1].(*Literal); ok {
		if lStr, ok := lt.Value().(string); ok && IsLegalKeyWord(lStr) {
			return fmt.Sprintf("%s.%s", i.params[0].convertToExpression(), i.params[0].convertToExpression())
		}
	}
	return fmt.Sprintf("%s[%s]", i.params[0].convertToExpression(), i.params[1].convertToExpression())
}

func (i *Index) convertToRealizedExpression(eCtx *EvaluationContext) string {
	exist, result := eCtx.tryGetTraceResult(i)
	if exist {
		return result
	}
	return fmt.Sprintf("%s[%s]", i.params[0].convertToExpression(), i.params[1].convertToExpression())
}

func (i *Index) evaluateCore(eCtx *EvaluationContext) any {
	l := evaluate(eCtx, i.params[0])
	isCol, col := l.TryGetCollectionInterface()
	if !isCol {
		_, isW := i.params[1].(*WildCard)
		if isW {
			return newFilteredArray()
		}
		return nil
	}
	fa, isFilteredArray := col.(*FilteredArray)
	if isFilteredArray {
		return i.handleFilteredArray(eCtx, fa)
	}
	obj, isObj := col.(IReadOnlyObj)
	if isObj {
		return i.handleObject(eCtx, obj)
	}
	arr, isArr := col.(IReadOnlyArray)
	if isArr {
		return i.handleArray(eCtx, arr)
	}
	return nil
}

func (i *Index) handleFilteredArray(eCtx *EvaluationContext, fa *FilteredArray) any {
	result := &FilteredArray{}
	idx := newIndexHelper(eCtx, i.params[1])
	ef := fa.Enumerator()
	for ef.Next() {
		item := ef.Value()
		itemResult := CreateIntermediateResult(eCtx, item)
		ok, nestedCollection := itemResult.TryGetCollectionInterface()
		if ok {
			if nestedObj, ok := nestedCollection.(IReadOnlyObj); ok {
				if idx.IsWildcard() {
					of := nestedObj.Enumerator()
					for of.Next() {
						result.Add(of.Value())
					}
				}
				if idx.HasStringIndex() {
					exist, nestedObjVal := nestedObj.GetValue(idx.StringIndex())
					if exist {
						result.Add(nestedObjVal)
					}
				}
			}
		}
		if nestedArray, ok := nestedCollection.(IReadOnlyArray); ok {
			if idx.IsWildcard() {
				af := nestedArray.Enumerator()
				for af.Next() {
					result.Add(af.Value())
				}
			}
			if idx.HasIntegerIndex() && idx.IntegerIndex() < nestedArray.Count() {
				result.Add(nestedArray.GetValue(idx.IntegerIndex()))
			}
		}

	}
	return result
}

func (i *Index) handleArray(eCtx *EvaluationContext, arr IReadOnlyArray) any {
	idx := newIndexHelper(eCtx, i.params[1])
	if idx.IsWildcard() {
		fa := newFilteredArray()
		e := arr.Enumerator()
		for e.Next() {
			fa.Add(e.Value())
		}
		return fa
	}
	if idx.HasIntegerIndex() && idx.IntegerIndex() < arr.Count() {
		return arr.GetValue(idx.IntegerIndex())
	}
	return nil
}

func (i *Index) handleObject(eCtx *EvaluationContext, obj IReadOnlyObj) any {
	idx := newIndexHelper(eCtx, i.params[1])
	if idx.IsWildcard() {
		fa := newFilteredArray()
		for _, v := range obj.Values() {
			fa.Add(v)
		}
		return fa
	}
	if idx.HasStringIndex() {
		exist, result := obj.GetValue(idx.StringIndex())
		if exist {
			return result
		}
	}
	return nil
}

func (e *Index) getContainer() iContainer {
	return e.container
}

func (e *Index) setContainer(c iContainer) {
	e.container = c
}

func (e *Index) getLevel() (level int) {
	return e.level
}

func (e *Index) getName() string {
	return e.name
}
func (e *Index) setName(name string) {
	e.name = name
}
