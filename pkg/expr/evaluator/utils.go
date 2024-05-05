package evaluator

import (
	"github.com/dungdm93/drasi/pkg/expr"
	"github.com/dungdm93/drasi/pkg/expr/common"
	"github.com/dungdm93/drasi/pkg/secret_masker"
)

func formatValueFromResult(masker secret_masker.Interface, result expr.Result) string {
	return common.FormatValue(masker, result.Value(), result.Kind())
}

func isPrimitive(kind expr.ResultKind) bool {
	switch kind {
	case expr.Null, expr.Boolean, expr.Number, expr.String:
		return true
	default:
		return false
	}
}
