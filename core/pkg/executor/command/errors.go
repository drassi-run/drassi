package command

import "errors"

var (
	ErrInvalidCommand       = errors.New("invalid command")
	ErrNotRegisteredCommand = errors.New("not registered command")
)
