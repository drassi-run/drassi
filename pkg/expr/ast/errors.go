package ast

import (
	"errors"
)

var (
	ErrorTooFewParameters  = errors.New("too few parameter")
	ErrorTooManyParameters = errors.New("too many parameter")
	ErrorMaxDepthExceeded  = errors.New("exceeded max depth")
	ErrorMaxLengthExceeded = errors.New("exceed max length")
)
