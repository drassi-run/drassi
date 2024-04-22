package parser

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression/constants"
	"github.com/dungdm93/drasi/pkg/expression/interfaces"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type Literal struct {
	base.ExpressionNodeBs
	value any
	kind  interfaces.ValueKind
	name  string
}

func NewLiteral(val any) *Literal {
	value, kind, _ := convertToCanonicalValue(val)
	return &Literal{
		value: value,
		kind:  kind,
		name:  kind.ToString(),
	}
}

func (l *Literal) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitLiteral(eCtx, l)
}

func (l *Literal) Value() any {
	return l.value
}

func (l *Literal) TraceFullyRealized() bool {
	return false
}

// func (l *Literal) EvaluateCore(eCtx interfaces.IEvaluationContext) any {
// 	return l.value
// }

func (l *Literal) ConvertToExpression() string {
	return formatValue(nil, l.value, l.kind)
}

func (l *Literal) ConvertToRealizedExpression(eCtx interfaces.IEvaluationContext) string {
	return formatValue(nil, l.value, l.kind)
}

func (l *Literal) GetContainer() interfaces.IContainer {
	return l.Container
}

func (l *Literal) GetLevel() (level int) {
	return l.Level
}

func (l *Literal) GetName() string {
	return l.name
}

func (l *Literal) SetName(name string) {
	l.name = name
}

// TODO: merge with evaluator.FormatValue
func formatValue(masker interfaces.ISecretMasker, value any, kind interfaces.ValueKind) string {
	switch kind {
	case interfaces.Null:
		return constants.Null
	case interfaces.Boolean:
		if value.(bool) {
			return constants.True
		}
		return constants.False
	case interfaces.Number:
		str := fmt.Sprintf("%f", value.(float64))
		if masker != nil {
			return masker.MaskSecrets(str)
		}
		return str
	case interfaces.String:
		str := value.(string)
		if masker != nil {
			str = masker.MaskSecrets(str)
		}
		return escapeSingleQuotes(str)
	case interfaces.Array, interfaces.Object:
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
func convertToCanonicalValue(input any) (value any, kind interfaces.ValueKind, raw any) {
	if input == nil {
		kind = interfaces.Null
		return
	}
	if _, castable := input.(bool); castable {
		kind = interfaces.Boolean
		value = input
		return
	}
	if _, castable := input.(float64); castable {
		value = input
		kind = interfaces.Number
		return
	}
	if _, castable := input.(string); castable {
		kind = interfaces.String
		value = input
		return
	}
	if b, castable := input.(interfaces.IBool); castable {
		kind = interfaces.Boolean
		raw = input
		value = b.GetValue()
		return
	}
	if b, castable := input.(interfaces.INumber); castable {
		kind = interfaces.Number
		raw = input
		value = b.GetValue()
		return
	}
	if b, castable := input.(interfaces.IString); castable {
		kind = interfaces.String
		raw = input
		value = b.GetValue()
		return
	}
	if _, castable := input.(interfaces.IReadOnlyObj); castable {
		kind = interfaces.Object
		value = input
		return
	}
	if _, castable := input.(interfaces.IReadOnlyArray); castable {
		kind = interfaces.Array
		value = input
		return
	}
	if _, castable := input.(interfaces.INull); castable {
		kind = interfaces.Null
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
			kind = interfaces.Number
			value, err := strconv.ParseFloat(input.(string), 64)
			if err != nil {
				panic(err)
			}
			return value, kind, nil
		}
	}
	kind = interfaces.Object
	value = input
	return
}
