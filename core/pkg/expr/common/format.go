package common

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dungdm93/drassi/core/pkg/expr"
	"github.com/dungdm93/drassi/core/pkg/secret_masker"
)

// ParseInt use strconv.ParseInt, allowing conversion from different bases (e.g., binary, octal, hexadecimal),
// max bitSize: 32. Int with value larger than 2147483647 is evaluated to float64
func ParseInt(str string) (out int, err error) {
	if len(str) == 0 || len(strings.TrimSpace(str)) == 0 {
		return 0, nil
	}
	i, err := strconv.ParseInt(str, 0, 32)
	if err == nil {
		return int(i), nil
	}
	return out, err
}

func ParseFloat(str string) (out float64) {
	if len(str) == 0 || len(strings.TrimSpace(str)) == 0 {
		return 0
	}
	out, err := strconv.ParseFloat(str, 64)
	if err == nil {
		return
	}
	// hex
	if str[0] == '0' && str[1] == 'x' && len(str) > 2 {
		for i := 1; i < len(str); i++ {
			x := str[i]
			if (x >= '0' && x <= '9') || (x >= 'a' && x <= 'f') || (x >= 'A' && x <= 'F') {
				// example:
				// Convert hexadecimal string to uint64
				if intVal, err := strconv.ParseUint(str[2:], 16, 64); err == nil {
					// Convert uint64 to float64
					return float64(intVal)
				}
			}
		}
	}
	if strings.EqualFold(str, Infinity) {
		return math.Inf(1)
	}
	if strings.EqualFold(str, NegativeInfinity) {
		return math.Inf(0)
	}
	return math.NaN()
}

func FormatValue(masker secret_masker.Interface, value any, kind expr.ResultKind) string {
	switch kind {
	case expr.Null:
		return Null
	case expr.Boolean:
		if value.(bool) {
			return True
		}
		return False
	case expr.Number:
		switch value.(type) {
		case float64:
			// without trailing zeros
			str := fmt.Sprintf("%.15g", value.(float64))
			if masker != nil {
				return masker.MaskSecrets(str)
			}
			return str
		case int:
			str := fmt.Sprintf("%d", value.(int))
			if masker != nil {
				return masker.MaskSecrets(str)
			}
			return str
		default:
			panic("FormatValue should not reach here")
		}
	case expr.String:
		str := value.(string)
		if masker != nil {
			str = masker.MaskSecrets(str)
		}
		return escapeSingleQuotes(str)
	case expr.Array, expr.Object:
		return kind.String()
	default:
		panic(fmt.Errorf("unable to format value. Unexpected value kind: %s", kind.String()))
	}
}

func escapeSingleQuotes(value string) string {
	if value == "" {
		return ""
	}
	return strings.ReplaceAll(value, "'", "''")
}

func ToCanonicalValue(input any) (value any, kind expr.ResultKind) {
	switch input.(type) {
	case nil:
		kind = expr.Null
		return
	case bool:
		kind = expr.Boolean
		value = input
		return
	case float64:
		// input is already float64
		kind = expr.Number
		f := input.(float64)
		floored := int(f)
		if float64(floored) == input && floored <= MaxInt32 && floored >= MinInt32 {
			value = floored
			return
		}
		value = input
		return
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32:
		// others well-known number
		// var floatType = reflect.TypeOf(float64(0))
		// kind = expr.Number
		// v := reflect.ValueOf(input)
		// v = reflect.Indirect(v)
		// fv := v.Convert(floatType)
		// return fv.Float(), kind
		kind = expr.Number
		value = input
		return
	case string:
		kind = expr.String
		value = input
		return
	case Obj:
		kind = expr.Object
		value = input
		return
	case Array:
		kind = expr.Array
		value = input
		return
	default:
		kind = expr.Object
		value = input
		return
	}
}
