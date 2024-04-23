package parser

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
	"github.com/dungdm93/drasi/pkg/secret_masker"
)

type literal struct {
	base.ExpressionNodeBs
	value any
	kind  expression.ValueKind
	name  string
}

func newLiteral(val any) *literal {
	value, kind, _ := convertToCanonicalValue(val)
	return &literal{
		value: value,
		kind:  kind,
		name:  kind.ToString(),
	}
}

func (l *literal) Accept(eCtx expression.IEvaluationContext, v expression.IExpressionNodeVisitor) any {
	return v.VisitLiteral(eCtx, l)
}

func (l *literal) Value() any {
	return l.value
}

func (l *literal) TraceFullyRealized() bool {
	return false
}

func (l *literal) ConvertToExpression() string {
	return formatValue(nil, l.value, l.kind)
}

func (l *literal) ConvertToRealizedExpression(eCtx expression.IEvaluationContext) string {
	return formatValue(nil, l.value, l.kind)
}

func (l *literal) GetContainer() expression.IContainer {
	return l.Container
}

func (l *literal) SetContainer(c expression.IContainer) {
	l.Container = c
}

func (l *literal) GetLevel() (level int) {
	return l.Level
}

func (l *literal) GetName() string {
	return l.name
}

func (l *literal) SetName(name string) {
	l.name = name
}

// TODO: merge with evaluator.FormatValue
func formatValue(masker interfaces.ISecretMasker, value any, kind expression.ValueKind) string {
	switch kind {
	case expression.ValueKindNull:
		return expression.Null
	case expression.ValueKindBoolean:
		if value.(bool) {
			return expression.True
		}
		return expression.False
	case expression.ValueKindNumber:
		str := fmt.Sprintf("%f", value.(float64))
		if masker != nil {
			return masker.MaskSecrets(str)
		}
		return str
	case expression.ValueKindString:
		str := value.(string)
		if masker != nil {
			str = masker.MaskSecrets(str)
		}
		return escapeSingleQuotes(str)
	case expression.ValueKindArray, expression.ValueKindObject:
		return kind.ToString()
	default:
		panic(fmt.Errorf("unable to convert to realized expression. Unexpected value kind: %s", kind))
	}
}

func escapeSingleQuotes(value string) string {
	if value == "" {
		return ""
	}
	return strings.ReplaceAll(value, "'", "''")
}

// TODO: merge with evaluator.ConvertToCanonicalValue
func convertToCanonicalValue(input any) (value any, kind expression.ValueKind, raw any) {
	if input == nil {
		kind = expression.ValueKindNull
		return
	}
	if _, castable := input.(bool); castable {
		kind = expression.ValueKindBoolean
		value = input
		return
	}
	if _, castable := input.(float64); castable {
		value = input
		kind = expression.ValueKindNumber
		return
	}
	if _, castable := input.(string); castable {
		kind = expression.ValueKindString
		value = input
		return
	}
	if b, castable := input.(expression.IBool); castable {
		kind = expression.ValueKindBoolean
		raw = input
		value = b.GetValue()
		return
	}
	if b, castable := input.(expression.INumber); castable {
		kind = expression.ValueKindNumber
		raw = input
		value = b.GetValue()
		return
	}
	if b, castable := input.(expression.IString); castable {
		kind = expression.ValueKindString
		raw = input
		value = b.GetValue()
		return
	}
	if _, castable := input.(expression.IReadOnlyObj); castable {
		kind = expression.ValueKindObject
		value = input
		return
	}
	if _, castable := input.(expression.IReadOnlyArray); castable {
		kind = expression.ValueKindArray
		value = input
		return
	}
	if _, castable := input.(expression.INull); castable {
		kind = expression.ValueKindNull
		raw = input
		value = nil
		return
	}
	if !(reflect.TypeOf(input).Kind() == reflect.Struct) {
		_, isInt := input.(int)
		_, isInt8 := input.(int8)
		_, isInt16 := input.(int16)
		_, isInt32 := input.(int32)
		_, isInt64 := input.(int64)

		_, isUint := input.(uint)
		_, isUint8 := input.(uint8)
		_, isUint16 := input.(uint16)
		_, isUint32 := input.(uint32)
		_, isUint64 := input.(uint64)

		_, isFloat32 := input.(float32)
		var isWellKnownNumber bool
		if isInt || isInt8 || isInt16 || isInt32 || isInt64 {
			isWellKnownNumber = true
		}
		if isUint || isUint8 || isUint16 || isUint32 || isUint64 {
			isWellKnownNumber = true
		}
		if isFloat32 {
			isWellKnownNumber = true
		}
		if isWellKnownNumber {
			kind = expression.ValueKindNumber
			value, err := strconv.ParseFloat(input.(string), 64)
			if err != nil {
				panic(err)
			}
			return value, kind, nil
		}
	}
	kind = expression.ValueKindObject
	value = input
	return
}
