package parser

// associativity represents the associativity of operators.
type associativity int

const (
	// associativityNone represents no associativity.
	associativityNone associativity = iota

	// associativityLTR represents left-to-right associativity.
	associativityLTR

	// associativityRTL represents right-to-left associativity.
	associativityRTL
)

func (a associativity) String() string {
	switch a {
	case associativityNone:
		return "None"
	case associativityLTR:
		return "LTR"
	case associativityRTL:
		return "RTL"
	default:
		return "Unknown associativity"
	}
}
