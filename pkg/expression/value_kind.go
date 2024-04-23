package expression

import (
	"fmt"
)

type ValueKind int

const (
	ValueKindArray ValueKind = iota
	ValueKindBoolean
	ValueKindNull
	ValueKindNumber
	ValueKindObject
	ValueKindString
)

func (v ValueKind) String() string {
	switch v {
	case ValueKindArray:
		return "ValueKindArray"
	case ValueKindBoolean:
		return "ValueKindBoolean"
	case ValueKindNull:
		return "ValueKindNull"
	case ValueKindNumber:
		return "ValueKindNumber"
	case ValueKindString:
		return "ValueKindString"
	case ValueKindObject:
		return "ValueKindObject"
	default:
		return "Unknown ValueKind"
	}
}

func (v ValueKind) ToString() string {
	return fmt.Sprintf("%s", v)
}
