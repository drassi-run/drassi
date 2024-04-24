package interfaces

type Visitor interface {
	VisitAlwaysFn(eCtx Context, c Node) any
	VisitAnd(eCtx Context, c Container) any
	VisitCancelledFn(eCtx Context, c Node) any
	VisitContainsFn(eCtx Context, c Container) any
	VisitContextValueNode(eCtx Context, c Node) any
	VisitEndsWithFn(eCtx Context, c Container) any
	VisitEqual(eCtx Context, c Container) any
	VisitFailureFn(eCtx Context, c Node) any
	VisitFormatFn(eCtx Context, c Container) any
	VisitFromJsonFn(eCtx Context, c Container) any
	VisitGreaterThan(eCtx Context, c Container) any
	VisitGreaterThanOrEqual(eCtx Context, c Container) any
	VisitIndex(eCtx Context, c Container) any
	VisitJoinFn(eCtx Context, c Container) any
	VisitLessThan(eCtx Context, c Container) any
	VisitLessThanOrEqual(eCtx Context, c Container) any
	VisitLiteral(eCtx Context, c Node) any
	VisitNoopFn(eCtx Context, c Node) any
	VisitNoopNamedValue(eCtx Context, c Node) any
	VisitNot(eCtx Context, c Container) any
	VisitNotEqual(eCtx Context, c Container) any
	VisitOr(eCtx Context, c Container) any
	VisitStartsWithFn(eCtx Context, c Container) any
	VisitSuccessFn(eCtx Context, c Node) any
	VisitWildCard(eCtx Context, c Node) any
}
