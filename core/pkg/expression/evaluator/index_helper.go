package evaluator

import (
	"math"

	"drassi.run/core/pkg/expression/ast"
	"drassi.run/core/pkg/expression/ast/ast_ifaces"
	"drassi.run/core/pkg/expression/common"
)

type (
	indexHelper struct {
		param  ast_ifaces.ExprNode
		result *result
		intIdx *int
		strIdx *string
	}
)

func newIndexHelper(eCtx ast_ifaces.Context, param ast_ifaces.ExprNode) *indexHelper {
	h := &indexHelper{
		param:  param,
		result: evaluate(eCtx, param),
	}

	h.intIdx = h.lazyIntegerIndex()
	h.strIdx = h.lazyStringIndex()
	return h
}

func (h *indexHelper) lazyStringIndex() *string {
	if h.result.primitive() {
		str := h.result.string()
		return &str
	}
	return nil
}

func (h *indexHelper) lazyIntegerIndex() *int {
	doubleIndex := h.result.number()
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
	_, ok := h.param.(*ast.WildCard)
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
	filteredArray = common.Array
)

func newFilteredArray() filteredArray {
	return make([]any, 0)
}

func handleFilteredArray(eCtx ast_ifaces.Context, fa filteredArray, i ast_ifaces.Container) any {
	result := filteredArray{}
	idx := newIndexHelper(eCtx, i.Parameters()[1])
	for _, value := range fa {
		item := value
		itemResult := createIntermediateResult(eCtx, item)
		ok, nestedCollection := itemResult.isCollection()
		if ok {
			if nestedObj, ok := nestedCollection.(common.Obj); ok {
				if idx.isWildcard() {
					for _, objValue := range nestedObj {
						result = append(result, objValue)
					}
				}
				if idx.hasStringIndex() {
					nestedObjVal, exist := nestedObj[idx.strIndex()]
					if exist {
						result = append(result, nestedObjVal)
					}
				}
			}
		}
		if nestedArray, ok := nestedCollection.(common.Array); ok {
			if idx.isWildcard() {
				for _, value := range nestedArray {
					result = append(result, value)
				}
			}
			if idx.hasIntegerIndex() && idx.intIndex() < len(nestedArray) {
				result = append(result, nestedArray[idx.intIndex()])
			}
		}
	}
	return result
}

func handleObject(eCtx ast_ifaces.Context, obj common.Obj, i ast_ifaces.Container) any {
	idx := newIndexHelper(eCtx, i.Parameters()[1])
	if idx.isWildcard() {
		fa := newFilteredArray()
		for _, v := range obj {
			fa = append(fa, v)
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

func handleArray(eCtx ast_ifaces.Context, arr common.Array, i ast_ifaces.Container) any {
	idx := newIndexHelper(eCtx, i.Parameters()[1])
	if idx.isWildcard() {
		fa := newFilteredArray()
		for _, value := range arr {
			fa = append(fa, value)
		}
		return fa
	}
	if idx.hasIntegerIndex() && idx.intIndex() < len(arr) {
		return arr[idx.intIndex()]
	}
	return nil
}
