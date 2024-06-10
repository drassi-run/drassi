package evaluator

import (
	"errors"
)

var (
	ErrorEmptyTrace               = errors.New("traceWriter must be provided")
	ErrorEmptySecretMasker        = errors.New("secret secretMasker must be provider")
	ErrorNonRootEvaluate          = errors.New("evaluate can only be called from root node")
	ErrorExprContextNotFound      = errors.New("expression context not found")
	ErrorExecutionContextNotFound = errors.New("execution context not found")
	ErrorInvalidFormatArgIndex    = errors.New("invalid format argument index")
)
