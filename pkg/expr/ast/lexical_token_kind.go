package ast

// tokenKind represents the k of token in the lexer.
type tokenKind int

const (
	// notInitialized is used when token is nil
	notInitialized tokenKind = iota
	// punctuation
	startGroup      // ( logical grouping
	startIndex      // [
	startParameters // ( function call
	endGroup        // ) logical grouping
	endIndex        // ]
	endParameters   // ) function call
	separator       // ,
	dereference     // .
	logicalOperator // !, ==, !=, >=, >, <, <=
	wildcard        // *
	// values
	null
	boolean
	number
	str
	propertyName
	function
	namedValue
	unexpected
)
