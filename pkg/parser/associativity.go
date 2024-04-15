package parser

// Associativity represents the associativity of operators.
type Associativity int

const (
	// AssociativityNone represents no associativity.
	AssociativityNone Associativity = iota

	// AssociativityLeftToRight represents left-to-right associativity.
	AssociativityLeftToRight

	// AssociativityRightToLeft represents right-to-left associativity.
	AssociativityRightToLeft
)

func (a Associativity) String() string {
	switch a {
	case AssociativityNone:
		return "none"
	case AssociativityLeftToRight:
		return "leftToRight"
	case AssociativityRightToLeft:
		return "rightToLeft"
	default:
		return "Unknown Associativity"
	}
}
