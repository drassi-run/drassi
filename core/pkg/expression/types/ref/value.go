package ref

// LazyVal is useful for short-circuit evaluation, e.g in `&&`, `||` operators
type LazyVal = func() Val

type Val interface {
	Type() Type

	// Value returns the raw value of the instance
	Value() any

	// Equal returns true if the `other` value has the same type and content as the implementing struct.
	Equal(other Val) bool
}

func IsError(v Val) bool {
	return v.Type() == TypeInvalid
}

func IsNull(v Val) bool {
	return v.Type() == TypeNull
}

func IsScalar(v Val) bool {
	switch v.Type() {
	case TypeNull, TypeBoolean, TypeInteger, TypeFloat, TypeString:
		return true
	default:
		return false
	}
}

func IsCollection(v Val) bool {
	switch v.Type() {
	case TypeList, TypeMap, TypeStruct:
		return true
	default:
		return false
	}
}
