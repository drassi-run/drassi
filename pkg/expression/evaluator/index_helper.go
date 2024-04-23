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
	filteredArray expression.ReadOnlyArray
)

func newFilteredArray() *filteredArray {
	a := make([]any, 0)
	return &filteredArray{a}
}

func (f *filteredArray) Add(v any) {
	*f = append(*f, v)
}

func (f *filteredArray) Count() int {
	return len(*f)
}

func (f *filteredArray) GetValue(idx int) any {
	return (*f)[idx]
}

func handleFilteredArray(eCtx expression.IEvaluationContext, fa *filteredArray, i expression.IContainer) any {
	result := &filteredArray{}
	idx := newIndexHelper(eCtx, i.Parameters()[1])
	for _, value := range *fa {
		item := value
		itemResult := createIntermediateResult(eCtx, item)
		ok, nestedCollection := itemResult.TryGetCollectionInterface()
		if ok {
			if nestedObj, ok := nestedCollection.(expression.ReadOnlyObj); ok {
				if idx.isWildcard() {
					for _, objValue := range nestedObj {
						result.Add(objValue)
					}
				}
				if idx.hasStringIndex() {
					nestedObjVal, exist := nestedObj[idx.strIndex()]
					if exist {
						result.Add(nestedObjVal)
					}
				}
			}
		}
		if nestedArray, ok := nestedCollection.(expression.ReadOnlyArray); ok {
			if idx.isWildcard() {
				for _, value := range nestedArray {
					result.Add(value)
				}
			}
			if idx.hasIntegerIndex() && idx.intIndex() < len(nestedArray) {
				result.Add(nestedArray[idx.intIndex()])
			}
		}
	}
	return result
}

func handleObject(eCtx expression.IEvaluationContext, obj expression.ReadOnlyObj, i expression.IContainer) any {
	idx := newIndexHelper(eCtx, i.Parameters()[1])
	if idx.isWildcard() {
		fa := newFilteredArray()
		for _, v := range obj {
			fa.Add(v)
		}
		return fa
	}
	if idx.hasStringIndex() {
		result, exist := obj[idx.strIndex()]
		if exist {
			return result
		}
	}
	return nil
}

func handleArray(eCtx expression.IEvaluationContext, arr expression.ReadOnlyArray, i expression.IContainer) any {
	idx := newIndexHelper(eCtx, i.Parameters()[1])
	if idx.isWildcard() {
		fa := newFilteredArray()
		for _, value := range arr {
			fa.Add(value)
		}
		return fa
	}
	if idx.hasIntegerIndex() && idx.intIndex() < len(arr) {
		return arr[idx.intIndex()]
	}
	return nil
}
