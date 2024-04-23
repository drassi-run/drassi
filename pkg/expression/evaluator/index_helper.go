package evaluator

import (
	"math"

	"github.com/dungdm93/drasi/pkg/expression"
	"github.com/dungdm93/drasi/pkg/expression/parser"
)

type (
	indexHelper struct {
		param  expression.IExpressionNode
		result *EvaluationResult
		intIdx *int
		strIdx *string
	}
)

func newIndexHelper(eCtx expression.IEvaluationContext, param expression.IExpressionNode) *indexHelper {
	h := &indexHelper{
		param:  param,
		result: evaluateWithContext(eCtx, param),
	}

	h.intIdx = h.lazyIntegerIndex()
	h.strIdx = h.lazyStringIndex()
	return h
}

func (h *indexHelper) lazyStringIndex() *string {
	if h.result.IsPrimitive() {
		str := h.result.ConvertToString()
		return &str
	}
	return nil
}

func (h *indexHelper) lazyIntegerIndex() *int {
	doubleIndex := h.result.ConvertToNumber()
	if math.IsNaN(doubleIndex) || doubleIndex < 0 {
		return nil
	}

	floorIndex := math.Floor(doubleIndex)
	if floorIndex > math.MaxInt {
		return nil
	}
	i := int(floorIndex)
	return &i
}

func (h *indexHelper) hasIntegerIndex() bool {
	return h.intIdx != nil
}

func (h *indexHelper) hasStringIndex() bool {
	return h.strIdx != nil
}

func (h *indexHelper) isWildcard() bool {
	_, ok := h.param.(*parser.WildCard)
	return ok
}

func (h *indexHelper) intIndex() int {
	if h.intIdx == nil {
		return 0
	}
	return *h.intIdx
}

func (h *indexHelper) strIndex() string {
	if h.strIdx == nil {
		return ""
	}
	return *h.strIdx
}

type (
	filteredArray struct {
		a []any
	}
)

func newFilteredArray() *filteredArray {
	return &filteredArray{a: make([]any, 0)}
}

func (f *filteredArray) Add(v any) {
	f.a = append(f.a, v)
}

func (f *filteredArray) Count() int {
	return len(f.a)
}

func (f *filteredArray) GetValue(idx int) any {
	return f.a[idx]
}

func (f *filteredArray) Enumerator() *expression.Enumerator {
	return expression.NewEnumerator(f.a)
}

func handleFilteredArray(eCtx expression.IEvaluationContext, fa *filteredArray, i expression.IContainer) any {
	result := &filteredArray{}
	idx := newIndexHelper(eCtx, i.Parameters()[1])
	ef := fa.Enumerator()
	for ef.Next() {
		item := ef.Value()
		itemResult := createIntermediateResult(eCtx, item)
		ok, nestedCollection := itemResult.TryGetCollectionInterface()
		if ok {
			if nestedObj, ok := nestedCollection.(expression.IReadOnlyObj); ok {
				if idx.isWildcard() {
					of := nestedObj.Enumerator()
					for of.Next() {
						result.Add(of.Value())
					}
				}
				if idx.hasStringIndex() {
					exist, nestedObjVal := nestedObj.GetValue(idx.strIndex())
					if exist {
						result.Add(nestedObjVal)
					}
				}
			}
		}
		if nestedArray, ok := nestedCollection.(expression.IReadOnlyArray); ok {
			if idx.isWildcard() {
				af := nestedArray.Enumerator()
				for af.Next() {
					result.Add(af.Value())
				}
			}
			if idx.hasIntegerIndex() && idx.intIndex() < nestedArray.Count() {
				result.Add(nestedArray.GetValue(idx.intIndex()))
			}
		}

	}
	return result
}

func handleObject(eCtx expression.IEvaluationContext, obj expression.IReadOnlyObj, i expression.IContainer) any {
	idx := newIndexHelper(eCtx, i.Parameters()[1])
	if idx.isWildcard() {
		fa := newFilteredArray()
		for _, v := range obj.Values() {
			fa.Add(v)
		}
		return fa
	}
	if idx.hasStringIndex() {
		exist, result := obj.GetValue(idx.strIndex())
		if exist {
			return result
		}
	}
	return nil
}

func handleArray(eCtx expression.IEvaluationContext, arr expression.IReadOnlyArray, i expression.IContainer) any {
	idx := newIndexHelper(eCtx, i.Parameters()[1])
	if idx.isWildcard() {
		fa := newFilteredArray()
		e := arr.Enumerator()
		for e.Next() {
			fa.Add(e.Value())
		}
		return fa
	}
	if idx.hasIntegerIndex() && idx.intIndex() < arr.Count() {
		return arr.GetValue(idx.intIndex())
	}
	return nil
}
