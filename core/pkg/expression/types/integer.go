package types

import (
	"fmt"
	"strconv"

	"drassi.run/core/pkg/expression/types/ref"
)

type Integer int64

func (i Integer) Value() any {
	return int64(i)
}

func (i Integer) Equal(other ref.Val) bool {
	o, ok := other.(Integer)
	return ok && i == o
}

func (i Integer) ToBoolean() bool {
	return i != 0
}

func (i Integer) ToNumber() float64 {
	return float64(i)
}

func (i Integer) ToString() string {
	return strconv.FormatInt(int64(i), 10)
}

func (i Integer) Compare(other ref.Val) (int, error) {
	o, ok := other.(Integer)
	if !ok {
		return 0, fmt.Errorf("cannot compare non-integer types")
	}

	if i < o {
		return -1, nil
	} else if i > o {
		return 1, nil
	} else {
		return 0, nil
	}
}
