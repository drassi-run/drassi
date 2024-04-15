package parser

type IReadOnlyArray interface {
	Count() int
	GetValue(idx int) any
	Enumerator() *Enumerator
}
