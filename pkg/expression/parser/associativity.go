package parser

// Associativity represents the associativity of operators.
type Associativity int

const (
	// AssociativityNone represents no associativity.
	AssociativityNone Associativity = iota

	// AssociativityLTR represents left-to-right associativity.
	AssociativityLTR

	// AssociativityRTL represents right-to-left associativity.
	AssociativityRTL
)

func (a Associativity) String() string {
	switch a {
	case AssociativityNone:
		return "None"
	case AssociativityLTR:
		return "LTR"
	case AssociativityRTL:
		return "RTL"
	default:
		return "Unknown Associativity"
	}
}
