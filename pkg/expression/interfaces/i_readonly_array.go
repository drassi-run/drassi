package interfaces

import (
	"github.com/dungdm93/drasi/pkg/expression/shared"
)

type IReadOnlyArray interface {
	Count() int
	GetValue(idx int) any
	Enumerator() *shared.Enumerator
}
