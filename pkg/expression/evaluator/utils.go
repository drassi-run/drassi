package evaluator

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression/constants"
	"github.com/dungdm93/drasi/pkg/expression/interfaces"
)

func FormatValue(masker interfaces.ISecretMasker, value any, kind interfaces.ValueKind) string {
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

func FormatValueFromResult(masker interfaces.ISecretMasker, result interfaces.IEvaluationResult) string {
	return FormatValue(masker, result.Value(), result.GetKind())
}

func IsPrimitive(kind interfaces.ValueKind) bool {
	switch kind {
	case interfaces.Null, interfaces.Boolean, interfaces.Number, interfaces.String:
		return true
	default:
		return false
	}
}

func ParseNumber(str string) (out float64) {
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
	if strings.EqualFold(str, constants.Infinity) {
		return math.Inf(1)
	}
	if strings.EqualFold(str, constants.NegativeInfinity) {
		return math.Inf(0)
	}
	return math.NaN()
}
