package types

import (
	"errors"

	"drassi.run/core/pkg/expression/types/ref"
)

var errUncomparable = errors.New("uncomparable data types")
var errNaNCompare = errors.New("NaN values cannot be ordered")

func NativeToVal(e any) ref.Val {
	// TODO
	return nil
}
