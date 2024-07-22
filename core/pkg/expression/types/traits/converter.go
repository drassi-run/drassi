package traits

// Logical interface used by logical operators (&&, || and !)
// Type is not implement this interface always assumed as `false`
type Logical interface {
	ToBoolean() bool
}

// Numerical interface used by relational operators (<, <=, >, >=) and equality operators (==, !=)
// Type is not implement this interface always assumed as `NaN`
type Numerical interface {
	ToNumber() float64
}

type Stringable interface {
	ToString() string
}
