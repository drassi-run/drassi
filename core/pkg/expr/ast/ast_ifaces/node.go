package ast_ifaces

type ExprNode interface {
	Expr() string
	RealizedExpr(eCtx Context) string
	// TraceFullyRealized indicates whether the evaluation result should be stored on the context and used when the realized result is traced.
	TraceFullyRealized() bool
	SetCtn(c Container)
	GetCtn() Container
	SetName(name string)
	GetName() string
	GetLevel() (level int)
	// Accept redirect to the visit method in the Visitor corresponding to the caller's type
	Accept(eCtx Context, visitor Visitor) any
	// Value is only meaningful for Literal node
	Value() any
}
