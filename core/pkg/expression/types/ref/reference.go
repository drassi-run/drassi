package ref

type Type int

func (t Type) String() string {
	switch t {
	case TypeNull:
		return "null"
	case TypeBoolean:
		return "boolean"
	case TypeInteger:
		return "number"
	case TypeFloat:
		return "number"
	case TypeString:
		return "string"
	case TypeList:
		return "array"
	case TypeMap:
		return "object"
	case TypeStruct:
		return "object"
	default:
		return "<invalid>"
	}
}

const (
	TypeInvalid Type = iota
	TypeNull
	TypeBoolean
	TypeInteger
	TypeFloat
	TypeString
	TypeList
	TypeMap
	TypeStruct
)

type Val interface {
	Type() Type

	// Value returns the raw value of the instance
	Value() any

	// Equal returns true if the `other` value has the same type and content as the implementing struct.
	Equal(other Val) bool
}

// LazyVal is useful for short-circuit evaluation, e.g in `&&`, `||` operators
type LazyVal = func() Val
