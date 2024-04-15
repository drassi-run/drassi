package parser

import (
	"fmt"
)

type ValueKind int

const (
	Array ValueKind = iota
	Boolean
	Null
	Number
	Object
	String
)

func (v ValueKind) String() string {
	switch v {
	case Array:
		return "Array"
	case Boolean:
		return "Boolean"
	case Null:
		return "Null"
	case Number:
		return "Number"
	case String:
		return "String"
	case Object:
		return "Object"
	default:
		return "Unknown value kind"
	}
}

func (v ValueKind) ToString() string {
	return fmt.Sprintf("%s", v)
}
