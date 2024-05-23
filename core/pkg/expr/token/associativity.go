package token

// Associativity represents the Associativity of operators.
type Associativity int

const (
	// AssociativityNone represents no Associativity.
	AssociativityNone Associativity = iota

	// AssociativityLTR represents left-to-right Associativity.
	AssociativityLTR

	// AssociativityRTL represents right-to-left Associativity.
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
		return "Unknown"
	}
}
