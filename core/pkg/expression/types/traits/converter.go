package traits

// Logical is designed for converting a type to a boolean value.
// It is primarily used in logical operators (&&, ||, and !)
// If a type does not implement this interface, it is automatically considered as `true`.
type Logical interface {
	ToBoolean() bool
}

// Numerical is designed for converting a type to a float64 value.
// It is primarily used in comparison operators (==, !=, <, >...)
// If a type does not implement this interface, it is automatically considered as `NaN`.
type Numerical interface {
	ToNumber() float64
}

// Stringable is designed for converting a type to a string value.
// If a type does not implement this interface, its ref.Type will be used to display its value.
type Stringable interface {
	ToString() string
}
