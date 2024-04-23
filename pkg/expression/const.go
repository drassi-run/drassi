package expression

const (
	False            = "false"
	Infinity         = "Infinity"
	MaxDepth         = 50
	NaN              = "NaN"
	NegativeInfinity = "-Infinity"
	Null             = "null"
	True             = "true"

	// Punctuation

	StartGroup     = "(" // logical grouping
	StartIndex     = "["
	StartParameter = "(" // function call
	EndGroup       = ")" // logical grouping
	EndIndex       = "]"
	EndParameter   = ")" // function call
	Separator      = ","
	Dereference    = "."
	Wildcard       = "*"

	// Operators

	Not                = "!"
	NotEqual           = "!="
	GreaterThan        = ">"
	GreaterThanOrEqual = ">="
	LessThan           = "<"
	LessThanOrEqual    = "<="
	Equal              = "=="
	And                = "&&"
	Or                 = "||"
)
