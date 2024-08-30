package types

import (
	"fmt"
	"strconv"

	"drassi.run/core/pkg/expression/types/ref"
)

type Integer int64

const (
	ZERO = Integer(0)
	ONE  = Integer(1)
)

func (i Integer) Type() ref.Type {
	return ref.TypeInteger
}

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
		return 0, fmt.Errorf("%w: %s vs. %s", errUncomparable, i.Type(), other.Type())
	}

	if i < o {
		return -1, nil
	} else if i > o {
		return 1, nil
	} else {
		return 0, nil
	}
}
