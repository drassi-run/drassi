package parser

import (
	"math"

	"github.com/dungdm93/drasi/pkg/expression/evaluator"
	interfaces2 "github.com/dungdm93/drasi/pkg/expression/interfaces"
)

func newIndexHelper(eCtx interfaces2.IEvaluationContext, param interfaces2.IExpressionNode) *IndexHelper {
	h := &IndexHelper{
		param:  param,
		result: evaluator.EvaluateWithContext(eCtx, param),
	}

	h.intIdx = h.lazyIntegerIndex()
	h.strIdx = h.lazyStringIndex()
	return h
}

func (h *IndexHelper) lazyStringIndex() *string {
	if h.result.IsPrimitive() {
		str := h.result.ConvertToString()
		return &str
	}
	return nil
}

func (h *IndexHelper) lazyIntegerIndex() *int {
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

func (h *IndexHelper) HasIntegerIndex() bool {
	return h.intIdx != nil
}
func (h *IndexHelper) HasStringIndex() bool {
	return h.strIdx != nil
}

func (h *IndexHelper) IsWildcard() bool {
	_, ok := h.param.(*WildCard)
	return ok
}

func (h *IndexHelper) IntegerIndex() int {
	if h.intIdx == nil {
		return 0
	}
	return *h.intIdx
}

func (h *IndexHelper) StringIndex() string {
	if h.strIdx == nil {
		return ""
	}
	return *h.strIdx
}

type (
	FilteredArray struct {
		a []any
	}
)

func newFilteredArray() *FilteredArray {
	return &FilteredArray{a: make([]any, 0)}
}

func (f *FilteredArray) Add(v any) {
	f.a = append(f.a, v)
}

func (f *FilteredArray) Count() int {
	return len(f.a)
}

func (f *FilteredArray) GetValue(idx int) any {
	return f.a[idx]
}

func (f *FilteredArray) Enumerator() *interfaces2.Enumerator {
	return interfaces2.NewEnumerator(f.a)
}
