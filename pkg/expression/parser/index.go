package parser

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expression/interfaces"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type (
	Index struct {
		base.ExpressionNodeBs
		base.ContainerBs
	}

	// IndexHelper struct {
	// 	param  interfaces.IExpressionNode
	// 	result *evaluator.EvaluationResult
	// 	intIdx *int
	// 	strIdx *string
	// }
)

func (a *Index) Value() any {
	panic("not implemented")
}

func (a *Index) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitIndex(eCtx, a)
}

func (i *Index) TraceFullyRealized() bool {
	return true
}

func (i *Index) ConvertToExpression() string {
	// Verify if we can simplify the expression, we would rather return
	// github.sha then github['sha'] so we check if this is a simple case.
	if lt, ok := i.Params[1].(*Literal); ok {
		if lStr, ok := lt.Value().(string); ok && IsLegalKeyWord(lStr) {
			return fmt.Sprintf("%s.%s", i.Params[0].ConvertToExpression(), i.Params[0].ConvertToExpression())
		}
	}
	return fmt.Sprintf("%s[%s]", i.Params[0].ConvertToExpression(), i.Params[1].ConvertToExpression())
}

func (i *Index) ConvertToRealizedExpression(eCtx interfaces.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(i)
	if exist {
		return result
	}
	return fmt.Sprintf("%s[%s]", i.Params[0].ConvertToExpression(), i.Params[1].ConvertToExpression())
}

//
// func (i *Index) EvaluateCore(eCtx interfaces.IEvaluationContext) any {
// 	l := evaluator.EvaluateWithContext(eCtx, i.Params[0])
// 	isCol, col := l.TryGetCollectionInterface()
// 	if !isCol {
// 		_, isW := i.Params[1].(*WildCard)
// 		if isW {
// 			return newFilteredArray()
// 		}
// 		return nil
// 	}
// 	fa, isFilteredArray := col.(*FilteredArray)
// 	if isFilteredArray {
// 		return i.handleFilteredArray(eCtx, fa)
// 	}
// 	obj, isObj := col.(interfaces.IReadOnlyObj)
// 	if isObj {
// 		return i.handleObject(eCtx, obj)
// 	}
// 	arr, isArr := col.(interfaces.IReadOnlyArray)
// 	if isArr {
// 		return i.handleArray(eCtx, arr)
// 	}
// 	return nil
// }

// func (i *Index) handleFilteredArray(eCtx interfaces.IEvaluationContext, fa *FilteredArray) any {
// 	result := &FilteredArray{}
// 	idx := newIndexHelper(eCtx, i.Params[1])
// 	ef := fa.Enumerator()
// 	for ef.Next() {
// 		item := ef.Value()
// 		itemResult := evaluator.CreateIntermediateResult(eCtx, item)
// 		ok, nestedCollection := itemResult.TryGetCollectionInterface()
// 		if ok {
// 			if nestedObj, ok := nestedCollection.(interfaces.IReadOnlyObj); ok {
// 				if idx.IsWildcard() {
// 					of := nestedObj.Enumerator()
// 					for of.Next() {
// 						result.Add(of.Value())
// 					}
// 				}
// 				if idx.HasStringIndex() {
// 					exist, nestedObjVal := nestedObj.GetValue(idx.StringIndex())
// 					if exist {
// 						result.Add(nestedObjVal)
// 					}
// 				}
// 			}
// 		}
// 		if nestedArray, ok := nestedCollection.(interfaces.IReadOnlyArray); ok {
// 			if idx.IsWildcard() {
// 				af := nestedArray.Enumerator()
// 				for af.Next() {
// 					result.Add(af.Value())
// 				}
// 			}
// 			if idx.HasIntegerIndex() && idx.IntegerIndex() < nestedArray.Count() {
// 				result.Add(nestedArray.GetValue(idx.IntegerIndex()))
// 			}
// 		}
//
// 	}
// 	return result
// }
//
// func (i *Index) handleArray(eCtx interfaces.IEvaluationContext, arr interfaces.IReadOnlyArray) any {
// 	idx := newIndexHelper(eCtx, i.Params[1])
// 	if idx.IsWildcard() {
// 		fa := newFilteredArray()
// 		e := arr.Enumerator()
// 		for e.Next() {
// 			fa.Add(e.Value())
// 		}
// 		return fa
// 	}
// 	if idx.HasIntegerIndex() && idx.IntegerIndex() < arr.Count() {
// 		return arr.GetValue(idx.IntegerIndex())
// 	}
// 	return nil
// }
//
// func (i *Index) handleObject(eCtx interfaces.IEvaluationContext, obj interfaces.IReadOnlyObj) any {
// 	idx := newIndexHelper(eCtx, i.Params[1])
// 	if idx.IsWildcard() {
// 		fa := newFilteredArray()
// 		for _, v := range obj.Values() {
// 			fa.Add(v)
// 		}
// 		return fa
// 	}
// 	if idx.HasStringIndex() {
// 		exist, result := obj.GetValue(idx.StringIndex())
// 		if exist {
// 			return result
// 		}
// 	}
// 	return nil
// }

func (i *Index) GetContainer() interfaces.IContainer {
	return i.Container
}

func (i *Index) SetContainer(c interfaces.IContainer) {
	i.Container = c
}

func (i *Index) GetLevel() (level int) {
	return i.Level
}

func (i *Index) GetName() string {
	return i.Name
}

func (i *Index) SetName(name string) {
	i.Name = name
}
