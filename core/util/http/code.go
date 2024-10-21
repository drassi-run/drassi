package utilhttp

// IsInformational checks if status code is 1xx: Request received, continuing process.
func IsInformational(code int) bool {
	return code/100 == 1
}

// IsSuccess checks if status code is 2xx: The action was successfully received,
// understood, and accepted.
func IsSuccess(code int) bool {
	return code/100 == 2
}

// IsRedirection checks if status code is 3xx: Further action must be taken in order
// to complete the request.
func IsRedirection(code int) bool {
	return code/100 == 3
}

// IsClientError checks if status code is 4xx: The request contains bad syntax or
// cannot be fulfilled.
func IsClientError(code int) bool {
	return code/100 == 4
}

// IsServerError checks if status code is 5xx: The server failed to fulfill
// an apparently valid request.
func IsServerError(code int) bool {
	return code/100 == 5
}
