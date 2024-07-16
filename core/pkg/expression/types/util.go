package types

import (
	"errors"

	"drassi.run/core/pkg/expression/types/ref"
)

var errUncomparable = errors.New("uncomparable data types")

func NativeToVal(e any) ref.Val {
	// TODO
	return nil
}
