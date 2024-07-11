package ref

type Val interface {
	// Value returns the raw value of the instance
	Value() any

	// Equal returns true if the `other` value has the same type and content as the implementing struct.
	Equal(other Val) bool
}

// LazyVal is useful for short-circuit evaluation, e.g in `&&`, `||` operators
type LazyVal = func() Val
