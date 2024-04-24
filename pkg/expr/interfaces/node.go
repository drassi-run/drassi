package interfaces

type Node interface {
	Expr() string
	RealizedExpr(eCtx Context) string
	SetCtn(c Container)
	GetCtn() Container
	SetName(name string)
	GetName() string
	// TraceFullyRealized indicates whether the evaluation result should be stored on the context and used when the realized result is traced.
	TraceFullyRealized() bool
	GetLevel() (level int)
	// Accept redirect to the visit method in the Visitor corresponding to the caller's type
	Accept(eCtx Context, visitor Visitor) any
	// Value is only meaningful for Literal node
	Value() any
}
